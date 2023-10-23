// Copyright 2023 Guillaume Charbonnier
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package azure

import (
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"

	"github.com/google/go-cmp/cmp"
	"github.com/libdns/libdns"
)

func Test_generateRecordSetName(t *testing.T) {
	t.Run("name=\"\"", func(t *testing.T) {
		got := generateRecordSetName("", "example.com.")
		want := "@"
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
	t.Run("name=@", func(t *testing.T) {
		got := generateRecordSetName("@", "example.com.")
		want := "@"
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
	t.Run("name=test", func(t *testing.T) {
		got := generateRecordSetName("test", "example.com.")
		want := "test"
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
	t.Run("name=test.example.com", func(t *testing.T) {
		got := generateRecordSetName("test.example.com", "example.com.")
		want := "test"
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
	t.Run("name=test.example.com.", func(t *testing.T) {
		got := generateRecordSetName("test.example.com.", "example.com.")
		want := "test"
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
	t.Run("name=example.com.", func(t *testing.T) {
		got := generateRecordSetName("example.com.", "example.com.")
		want := "@"
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
}

func Test_convertStringToRecordType(t *testing.T) {
	typeNames := []string{"A", "AAAA", "CAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"}
	for _, typeName := range typeNames {
		t.Run("type="+typeName, func(t *testing.T) {
			recordType, _ := convertStringToRecordType(typeName)
			got := fmt.Sprintf("%T:%v", recordType, recordType)
			want := "armdns.RecordType:" + typeName
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
	t.Run("type=ERR", func(t *testing.T) {
		_, err := convertStringToRecordType("ERR")
		got := err.Error()
		want := "The type ERR cannot be interpreted."
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
}

func Test_convertAzureRecordSetsToLibdnsRecords(t *testing.T) {
	t.Run("type=supported", func(t *testing.T) {
		azureRecordSets := []*armdns.RecordSet{
			{
				Name: to.Ptr("record-a"),
				Type: to.Ptr("Microsoft.Network/dnszones/A"),
				Etag: to.Ptr("ETAG_A"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-a.example.com."),
					ARecords: []*armdns.ARecord{
						{
							IPv4Address: to.Ptr("127.0.0.1"),
						},
					},
				},
			},
			{
				Name: to.Ptr("record-aaaa"),
				Type: to.Ptr("Microsoft.Network/dnszones/AAAA"),
				Etag: to.Ptr("ETAG_AAAA"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-aaaa.example.com."),
					AaaaRecords: []*armdns.AaaaRecord{{
						IPv6Address: to.Ptr("::1"),
					}},
				},
			},
			{
				Name: to.Ptr("record-caa"),
				Type: to.Ptr("Microsoft.Network/dnszones/CAA"),
				Etag: to.Ptr("ETAG_CAA"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-caa.example.com."),
					CaaRecords: []*armdns.CaaRecord{{
						Flags: to.Ptr[int32](0),
						Tag:   to.Ptr("issue"),
						Value: to.Ptr("ca.example.com"),
					}},
				},
			},
			{
				Name: to.Ptr("record-cname"),
				Type: to.Ptr("Microsoft.Network/dnszones/CNAME"),
				Etag: to.Ptr("ETAG_CNAME"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-cname.example.com."),
					CnameRecord: &armdns.CnameRecord{
						Cname: to.Ptr("www.example.com"),
					},
				},
			},
			{
				Name: to.Ptr("record-mx"),
				Type: to.Ptr("Microsoft.Network/dnszones/MX"),
				Etag: to.Ptr("ETAG_MX"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-mx.example.com."),
					MxRecords: []*armdns.MxRecord{{
						Preference: to.Ptr[int32](10),
						Exchange:   to.Ptr("mail.example.com"),
					}},
				},
			},
			{
				Name: to.Ptr("@"),
				Type: to.Ptr("Microsoft.Network/dnszones/NS"),
				Etag: to.Ptr("ETAG_NS"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("example.com."),
					NsRecords: []*armdns.NsRecord{
						{
							Nsdname: to.Ptr("ns1.example.com"),
						},
						{
							Nsdname: to.Ptr("ns2.example.com"),
						},
					},
				},
			},
			{
				Name: to.Ptr("record-ptr"),
				Type: to.Ptr("Microsoft.Network/dnszones/PTR"),
				Etag: to.Ptr("ETAG_PTR"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-ptr.example.com."),
					PtrRecords: []*armdns.PtrRecord{{
						Ptrdname: to.Ptr("hoge.example.com"),
					}},
				},
			}, {
				Name: to.Ptr("@"),
				Type: to.Ptr("Microsoft.Network/dnszones/SOA"),
				Etag: to.Ptr("ETAG_SOA"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("example.com."),
					SoaRecord: &armdns.SoaRecord{
						Host:         to.Ptr("ns1.example.com"),
						Email:        to.Ptr("hostmaster.example.com"),
						SerialNumber: to.Ptr[int64](1),
						RefreshTime:  to.Ptr[int64](7200),
						RetryTime:    to.Ptr[int64](900),
						ExpireTime:   to.Ptr[int64](1209600),
						MinimumTTL:   to.Ptr[int64](86400),
					},
				},
			},
			{
				Name: to.Ptr("record-srv"),
				Type: to.Ptr("Microsoft.Network/dnszones/SRV"),
				Etag: to.Ptr("ETAG_SRV"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-srv.example.com."),
					SrvRecords: []*armdns.SrvRecord{{
						Priority: to.Ptr[int32](1),
						Weight:   to.Ptr[int32](10),
						Port:     to.Ptr[int32](5269),
						Target:   to.Ptr("app.example.com"),
					}},
				},
			},
			{
				Name: to.Ptr("record-txt"),
				Type: to.Ptr("Microsoft.Network/dnszones/TXT"),
				Etag: to.Ptr("ETAG_TXT"),
				Properties: &armdns.RecordSetProperties{
					TTL:  to.Ptr[int64](30),
					Fqdn: to.Ptr("record-txt.example.com."),
					TxtRecords: []*armdns.TxtRecord{{
						Value: []*string{to.Ptr("TEST VALUE")},
					}},
				},
			},
		}
		got, _ := convertAzureRecordSetsToLibdnsRecords(azureRecordSets)
		want := []libdns.Record{
			{
				ID:    "ETAG_A",
				Type:  "A",
				Name:  "record-a",
				Value: "127.0.0.1",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_AAAA",
				Type:  "AAAA",
				Name:  "record-aaaa",
				Value: "::1",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_CAA",
				Type:  "CAA",
				Name:  "record-caa",
				Value: "0 issue ca.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_CNAME",
				Type:  "CNAME",
				Name:  "record-cname",
				Value: "www.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_MX",
				Type:  "MX",
				Name:  "record-mx",
				Value: "10 mail.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_NS",
				Type:  "NS",
				Name:  "@",
				Value: "ns1.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_NS",
				Type:  "NS",
				Name:  "@",
				Value: "ns2.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_PTR",
				Type:  "PTR",
				Name:  "record-ptr",
				Value: "hoge.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_SOA",
				Type:  "SOA",
				Name:  "@",
				Value: "ns1.example.com hostmaster.example.com 1 7200 900 1209600 86400",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_SRV",
				Type:  "SRV",
				Name:  "record-srv",
				Value: "1 10 5269 app.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_TXT",
				Type:  "TXT",
				Name:  "record-txt",
				Value: "TEST VALUE",
				TTL:   time.Duration(30) * time.Second,
			},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
	t.Run("type=unsupported", func(t *testing.T) {
		azureRecordSets := []*armdns.RecordSet{{
			Type: to.Ptr("Microsoft.Network/dnszones/ERR"),
		}}
		_, err := convertAzureRecordSetsToLibdnsRecords(azureRecordSets)
		got := err.Error()
		want := "The type ERR cannot be interpreted."
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
}

func Test_convertLibdnsRecordToAzureRecordSet(t *testing.T) {
	t.Run("type=supported", func(t *testing.T) {
		libdnsRecords := []libdns.Record{
			{
				ID:    "ETAG_A",
				Type:  "A",
				Name:  "record-a",
				Value: "127.0.0.1",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_AAAA",
				Type:  "AAAA",
				Name:  "record-aaaa",
				Value: "::1",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_CAA",
				Type:  "CAA",
				Name:  "record-caa",
				Value: "0 issue ca.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_CNAME",
				Type:  "CNAME",
				Name:  "record-cname",
				Value: "www.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_MX",
				Type:  "MX",
				Name:  "record-mx",
				Value: "10 mail.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_NS",
				Type:  "NS",
				Name:  "@",
				Value: "ns1.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_PTR",
				Type:  "PTR",
				Name:  "record-ptr",
				Value: "hoge.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_SOA",
				Type:  "SOA",
				Name:  "@",
				Value: "ns1.example.com hostmaster.example.com 1 7200 900 1209600 86400",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_SRV",
				Type:  "SRV",
				Name:  "record-srv",
				Value: "1 10 5269 app.example.com",
				TTL:   time.Duration(30) * time.Second,
			},
			{
				ID:    "ETAG_TXT",
				Type:  "TXT",
				Name:  "record-txt",
				Value: "TEST VALUE",
				TTL:   time.Duration(30) * time.Second,
			},
		}
		var got []armdns.RecordSet
		for _, libdnsRecord := range libdnsRecords {
			convertedRecord, _ := convertLibdnsRecordToAzureRecordSet(libdnsRecord)
			got = append(got, convertedRecord)
		}
		want := []armdns.RecordSet{
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					ARecords: []*armdns.ARecord{{
						IPv4Address: to.Ptr("127.0.0.1"),
					}},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					AaaaRecords: []*armdns.AaaaRecord{{
						IPv6Address: to.Ptr("::1"),
					}},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					CaaRecords: []*armdns.CaaRecord{{
						Flags: to.Ptr[int32](0),
						Tag:   to.Ptr("issue"),
						Value: to.Ptr("ca.example.com"),
					}},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					CnameRecord: &armdns.CnameRecord{
						Cname: to.Ptr("www.example.com"),
					},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					MxRecords: []*armdns.MxRecord{{
						Preference: to.Ptr[int32](10),
						Exchange:   to.Ptr("mail.example.com"),
					}},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					NsRecords: []*armdns.NsRecord{{
						Nsdname: to.Ptr("ns1.example.com"),
					}},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					PtrRecords: []*armdns.PtrRecord{{
						Ptrdname: to.Ptr("hoge.example.com"),
					}},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					SoaRecord: &armdns.SoaRecord{
						Host:         to.Ptr("ns1.example.com"),
						Email:        to.Ptr("hostmaster.example.com"),
						SerialNumber: to.Ptr[int64](1),
						RefreshTime:  to.Ptr[int64](7200),
						RetryTime:    to.Ptr[int64](900),
						ExpireTime:   to.Ptr[int64](1209600),
						MinimumTTL:   to.Ptr[int64](86400),
					},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					SrvRecords: []*armdns.SrvRecord{{
						Priority: to.Ptr[int32](1),
						Weight:   to.Ptr[int32](10),
						Port:     to.Ptr[int32](5269),
						Target:   to.Ptr("app.example.com"),
					}},
				},
			},
			{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr[int64](30),
					TxtRecords: []*armdns.TxtRecord{{
						Value: []*string{to.Ptr("TEST VALUE")},
					}},
				},
			},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
	t.Run("type=unsupported", func(t *testing.T) {
		libdnsRecords := []libdns.Record{{
			Type: "ERR",
		}}
		_, err := convertLibdnsRecordToAzureRecordSet(libdnsRecords[0])
		got := err.Error()
		want := "The type ERR cannot be interpreted."
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("diff: %s", diff)
		}
	})
}
