package common_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/quara-dev/beyond/modules/nats/connectors/components/mongo/common"
)

var _ = Describe("Events", func() {
	Context("Generate write model", func() {
		It("should return a write model for insert operation", func() {
			raw := `{"documentKey": {"_id":  {"$oid": "5f3f9d9b2f0d3a0001c3c4f4"}}, "operationType": "insert", "fullDocument": {"foo": "bar"}}`
			change := common.ChangeStreamEvent{
				Raw: []byte(raw),
			}
			model, err := change.WriteModel()
			Expect(err).To(BeNil())
			Expect(model).ToNot(BeNil())
			Expect(model.(*mongo.InsertOneModel).Document).To(Equal(map[string]interface{}{"foo": "bar"}))
		})
		It("should return a write model for update operation", func() {
			raw := `{"documentKey": {"_id":  {"$oid": "5f3f9d9b2f0d3a0001c3c4f4"}}, "operationType": "update", "updateDescription": {"updatedFields": {"foo": "bar"}, "removedFields": ["bar"]}}`
			change := common.ChangeStreamEvent{
				Raw: []byte(raw),
			}
			model, err := change.WriteModel()
			Expect(err).To(BeNil())
			Expect(model).ToNot(BeNil())
			Expect(model.(*mongo.UpdateOneModel).Update).To(Equal(map[string]interface{}{"$set": map[string]interface{}{"foo": "bar"}, "$unset": []string{"bar"}}))
		})
		It("should return a write model for delete operation", func() {
			raw := `{"documentKey": {"_id":  {"$oid": "5f3f9d9b2f0d3a0001c3c4f4"}}, "operationType": "delete"}`
			change := common.ChangeStreamEvent{
				Raw: []byte(raw),
			}
			model, err := change.WriteModel()
			Expect(err).To(BeNil())
			Expect(model).ToNot(BeNil())
			obj := primitive.NewObjectID()
			obj.UnmarshalText([]byte("5f3f9d9b2f0d3a0001c3c4f4"))
			Expect(model.(*mongo.DeleteOneModel).Filter).To(Equal(map[string]interface{}{"_id": obj}))
		})
		It("should return an error", func() {
			raw := `{"documentKey": {"_id": 123}}`
			change := common.ChangeStreamEvent{
				Raw: []byte(raw),
			}
			_, err := change.WriteModel()
			Expect(err).ToNot(BeNil())
		})
	})
})
