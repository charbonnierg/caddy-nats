package common

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewChangeStreamEvent(raw []byte) (*ChangeStreamEvent, error) {
	change := &ChangeStreamEvent{
		Raw: raw,
	}
	_, err := change.LookupOperationType()
	if err != nil {
		return nil, err
	}
	_, err = change.lookupDatabase()
	if err != nil {
		return nil, err
	}
	_, err = change.LookupCollection()
	if err != nil {
		return nil, err
	}
	_, err = change.lookupDocumentId()
	if err != nil {
		return nil, err
	}
	return change, nil
}

type ChangeStreamEvent struct {
	raw           *bson.Raw
	collection    string
	database      string
	operationType string
	documentId    primitive.ObjectID

	Raw []byte
}

func (m *ChangeStreamEvent) lookup(key ...string) bson.RawValue {
	if m.raw != nil {
		return (*m.raw).Lookup(key...)
	}
	v := map[string]interface{}{}
	err := bson.UnmarshalExtJSON(m.Raw, false, &v)
	if err != nil {
		return bson.RawValue{}
	}
	raw, err := bson.Marshal(v)
	if err != nil {
		return bson.RawValue{}
	}
	rraw := bson.Raw(raw)
	m.raw = &rraw
	return rraw.Lookup(key...)
}

func (m *ChangeStreamEvent) lookupDocumentId() (primitive.ObjectID, error) {
	if m.documentId != primitive.NilObjectID {
		return m.documentId, nil
	}
	id, ok := m.lookup("documentKey", "_id").ObjectIDOK()
	if !ok {
		return primitive.NewObjectID(), errors.New("invalid change stream document: no document key")
	}
	m.documentId = id
	return id, nil
}

func (m *ChangeStreamEvent) LookupOperationType() (string, error) {
	if m.operationType != "" {
		return m.operationType, nil
	}
	op, ok := m.lookup("operationType").StringValueOK()
	if !ok {
		return "", errors.New("invalid change stream document: no operation type")
	}
	m.operationType = op
	return op, nil
}

func (m *ChangeStreamEvent) lookupDatabase() (string, error) {
	if m.database != "" {
		return m.database, nil
	}
	db, ok := m.lookup("ns", "db").StringValueOK()
	if !ok {
		return "", errors.New("invalid change stream document: no database")
	}
	m.database = db
	return db, nil
}

func (m *ChangeStreamEvent) LookupCollection() (string, error) {
	if m.collection != "" {
		return m.collection, nil
	}
	coll, ok := m.lookup("ns", "coll").StringValueOK()
	if !ok {
		return "", errors.New("invalid change stream document: no collection")
	}
	m.collection = coll
	return coll, nil
}

func (m *ChangeStreamEvent) lookupFullDocument() (interface{}, error) {
	doc := map[string]interface{}{}
	err := m.lookup("fullDocument").Unmarshal(&doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (m *ChangeStreamEvent) lookupUpdatedFields() (interface{}, error) {
	doc := map[string]interface{}{}
	err := m.lookup("updateDescription", "updatedFields").Unmarshal(&doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (m *ChangeStreamEvent) lookupRemovedFields() (interface{}, error) {
	doc := []string{}
	err := m.lookup("updateDescription", "removedFields").Unmarshal(&doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (m *ChangeStreamEvent) generateFilter() (interface{}, error) {
	id, err := m.lookupDocumentId()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"_id": id,
	}, nil
}

func (m *ChangeStreamEvent) generateUpdateDocument() (interface{}, error) {
	updated, err := m.lookupUpdatedFields()
	if err != nil {
		return nil, err
	}
	removed, err := m.lookupRemovedFields()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"$set":   updated,
		"$unset": removed,
	}, nil
}

func (m *ChangeStreamEvent) WriteModel() (mongo.WriteModel, error) {
	op, err := m.LookupOperationType()
	if err != nil {
		return nil, err
	}
	switch op {
	case "insert":
		fullDoc, err := m.lookupFullDocument()
		if err != nil {
			return nil, err
		}
		return &mongo.InsertOneModel{
			Document: fullDoc,
		}, nil
	case "update":
		updateDoc, err := m.generateUpdateDocument()
		if err != nil {
			return nil, err
		}
		updateFilter, err := m.generateFilter()
		if err != nil {
			return nil, err
		}
		return &mongo.UpdateOneModel{
			Update: updateDoc,
			Filter: updateFilter,
		}, nil
	case "delete":
		deleteFilter, err := m.generateFilter()
		if err != nil {
			return nil, err
		}
		return &mongo.DeleteOneModel{
			Filter: deleteFilter,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported operation: %s", op)
	}
}

func (m *ChangeStreamEvent) Payload() ([]byte, error) {
	return m.Raw, nil
}

func (m *ChangeStreamEvent) Subject(prefix string) (string, error) {
	db, err := m.lookupDatabase()
	if err != nil {
		return "", err
	}
	coll, err := m.LookupCollection()
	if err != nil {
		return "", err
	}
	op, err := m.LookupOperationType()
	if err != nil {
		return "", err
	}
	id, err := m.lookupDocumentId()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s.%s.%s.%s", prefix, db, coll, op, id.Hex()), nil
}

func (m *ChangeStreamEvent) Headers() (map[string][]string, error) {
	db, err := m.lookupDatabase()
	if err != nil {
		return nil, err
	}
	coll, err := m.LookupCollection()
	if err != nil {
		return nil, err
	}
	op, err := m.LookupOperationType()
	if err != nil {
		return nil, err
	}
	id, err := m.lookupDocumentId()
	if err != nil {
		return nil, err
	}
	return map[string][]string{
		"Mongo-OperationType": {op},
		"Mongo-DocumentId":    {id.Hex()},
		"Mongo-Collection":    {coll},
		"Mongo-Database":      {db},
	}, nil
}
