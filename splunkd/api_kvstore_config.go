package splunkd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/prigio/splunk-go-sdk/utils"
)

// This file provides structs used to parse the JSON-formatted output of the Splunk REST API

// See: https://docs.splunk.com/Documentation/Splunk/8.1.3/RESTREF/RESTprolog

const (
	KVStoreFieldTypeNone   string = ""
	KVStoreFieldTypeArray  string = "array"
	KVStoreFieldTypeNumber string = "number"
	KVStoreFieldTypeBool   string = "bool"
	KVStoreFieldTypeString string = "string"
	KVStoreFieldTypeCIDR   string = "cidr"
	KVStoreFieldTypeTime   string = "time"
)

type KVStoreFieldDefinition struct {
	Name string
	Type string
}

// KVStoreCollResource represents the definition of a KVStore collection
type KVStoreCollResource struct {
	Disabled     bool `json:"disabled"`
	EnforceTypes bool `json:"enforceTypes"`
	Replicate    bool `json:"replicate"`

	Fields            map[string]string
	AcceleratedFields map[string]string
}

// UnmarshalJSON implements the JSON custom unmarshaller interface to properly convert from the API JSON based results
// to the internal data structure
func (kvcr *KVStoreCollResource) UnmarshalJSON(data []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	kvcr.Disabled = interfaceToBool(tmp["disabled"])
	kvcr.EnforceTypes = interfaceToBool(tmp["enforceTypes"])
	kvcr.Replicate = interfaceToBool(tmp["replicate"])

	kvcr.Fields = make(map[string]string, 0)
	kvcr.AcceleratedFields = make(map[string]string)

	for k, v := range tmp {
		if strings.HasPrefix(k, "field.") {
			kvcr.Fields[strings.TrimPrefix(k, "field.")] = v.(string)
		} else if strings.HasPrefix(k, "accelerated_fields.") {
			kvcr.AcceleratedFields[strings.TrimPrefix(k, "accelerated_fields.")] = v.(string)
		}
	}
	return nil
}

/*
	Query retrieves data from the specifed collection

query - Query JSON object.
Conditional operators: $gt, $gte, $lt, $lte, and $ne
Logical operators: $and, $or, and ,$not (invert conditional operators)
Examples:
query="{}" (select all documents)
query={"title":"Item"} (Select all documents with property title that has value Item)
query={"price":{"$gt":5}} (Select all documents with price greater than 5)
fields - Comma-separated list of fields to include (1) or exclude (0). A fields value cannot contain both include and exclude specifications except for exclusion of the _key field. Examples:
fields=firstname,surname (Include only firstname, surname, and _key fields)
fields=firstname,surname,_key:0 (Include only the firstname and surname fields)
fields=address:0 (Include all fields except the address field)
limit -  	Maximum number of items to return.
skip - Number of items to skip from the start.
sort - Sort order. Examples:
sort=surname (Sort by surname, ascending)
sort=surname,firstname (Sort by surname, ascending, after firstname, ascending)
sort=surname:-1,firstname:1 (Sort by surname, descending, after firstname, ascending
sort=surname:1,first name (Sort by surname, ascending, after firstname, ascending
shared - Defaults to false. Set to true to return records for the specified user as well as records for the nobody user.
*/
func (entry *collectionEntry[KVStoreCollResource]) Query(ss *Client, query, fields, sort string, limit, skip int, shared bool, storeJSONResultInto *[]map[string]interface{}) error {
	ctx := fmt.Sprintf("kvstore[%s] query", entry.Name)
	if ss == nil {
		return utils.NewErrInvalidParam(ctx, nil, "'splunkService' cannot be nil")
	}
	if query == "" {
		return utils.NewErrInvalidParam(ctx, nil, "'query' cannot be empty. Provide \"{}\" to select all documents")
	}
	if storeJSONResultInto == nil {
		return utils.NewErrInvalidParam(ctx, nil, "'storeJSONResultInto' cannot be nil")
	}

	dataURL := strings.ReplaceAll(entry.Links.List, "/collections/config/", "/collections/data/")
	queryParams := url.Values{}
	queryParams.Set("query", query)
	if sort != "" {
		queryParams.Set("sort", sort)
	}
	if limit > 0 {
		queryParams.Set("limit", strconv.FormatInt(int64(limit), 10))
	}
	if skip > 0 {
		queryParams.Set("skip", strconv.FormatInt(int64(skip), 10))
	}
	if err := doSplunkdHttpRequest(ss, "GET", dataURL, &queryParams, nil, "", storeJSONResultInto); err != nil {
		return fmt.Errorf("kvstore[%s] query: %w", entry.Name, err)
	}
	return nil
}

func (entry *collectionEntry[KVStoreCollResource]) Insert(ss *Client, jsondata string) (key string, err error) {
	ctx := fmt.Sprintf("kvstore[%s] insert", entry.Name)
	if ss == nil {
		return "", utils.NewErrInvalidParam(ctx, nil, "'splunkService' cannot be nil")
	}
	if jsondata == "" {
		return "", utils.NewErrInvalidParam(ctx, nil, "'jsondata' cannot be empty")
	}
	dataURL := strings.ReplaceAll(entry.Links.List, "/collections/config/", "/collections/data/")
	dataRes := make(map[string]string, 0)
	if err = doSplunkdHttpRequest(ss, "POST", dataURL, nil, []byte(jsondata), "application/json", &dataRes); err != nil {
		return "", fmt.Errorf("%s: %w", ctx, err)
	}
	return dataRes["_key"], nil
}

// KVStoreCollCollection represents a collection of definitions of KV Store collections as managed by the /services/storage/collections/config endpoint.
// This also supports custom configuration files defined with a custom SPEC file within etc/apps/<someapp>/README/<somefile>.conf.spec.
// See: https://docs.splunk.com/Documentation/Splunk/9.0.5/RESTREF/RESTkvstore#storage.2Fcollections.2Fconfig.2F.7Bcollection.7D
type KVStoreCollCollection struct {
	collection[KVStoreCollResource]
}

func NewKVStoreCollCollection(ss *Client) *KVStoreCollCollection {
	var col = &KVStoreCollCollection{}
	col.name = "KVStore collection"
	col.path = "/servicesNS/nobody/-/storage/collections/config"
	col.splunkd = ss
	return col
}

func (col *KVStoreCollCollection) CreateKVStoreColl(ns *Namespace, entryName string, fields map[string]string, acceleratedFields map[string]string, enforceTypes bool, replicate bool) (*collectionEntry[KVStoreCollResource], error) {
	params := url.Values{}
	params.Set("name", entryName)
	params.Set("replicate", fmt.Sprintf("%v", replicate))
	params.Set("enforceTypes", fmt.Sprintf("%v", enforceTypes))

	for fName, fType := range fields {
		params.Set("field."+fName, fType)
	}
	for k, v := range acceleratedFields {
		params.Set("accelerated_fields."+k, v)
	}
	//ns, _ := NewNamespace("nobody", "-", SplunkSharingApp)
	return col.CreateNS(ns, entryName, &params)
}
