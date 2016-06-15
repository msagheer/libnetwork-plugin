// Copyright 2016 PLUMgrid
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package main

package driver

import (
	"encoding/json"
)

func BridgeCreate(ID string, domainid string, gatewayip string) {

	url := "/0/connectivity/domain/" + domainid + "/ne/bri" + ID
	data := []byte(`{"action": {
				"action1": {
					"action_text": "create_and_link_ifc(DYN_1)"}
				},
				"ifc": {},
				"mobility": "true",
				"ne_description": "PLUMgrid Bridge",
				"ne_dname": "net-` + ID[:10] + `",
				"ne_group": "Bridge",
				"ne_type": "bridge"
	}`)

	RestCall("PUT", url, data)

	url = "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID
	data = []byte(`{"mark_disabled": false,
			"ne_dest": "/ne/bri` + ID + `/action/action1",
			"ne_dname": "cont-` + ID[10:] + `",
			"ne_type": "cnf-vmgroup",
					"rule": {
						"rules` + ID + `": {
							"add_context": "",
							"criteria": "pgtag2",
							"match": "bridge-` + ID[:10] + `"
						}
					}
	}`)

	RestCall("PUT", url, data)

	url = "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID
	data = []byte(`{"ne_metadata": "` + gatewayip + `"}`)
	RestCall("PUT", url, data)
}

func BridgeDelete(ID string, domainid string) {

	url := "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID
	RestCall("DELETE", url, nil)

	url = "/0/connectivity/domain/" + domainid + "/ne/bri" + ID
	RestCall("DELETE", url, nil)

	url = "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID
	RestCall("DELETE", url, nil)
}

func FindDomainFromNetwork(ID string) (domainid string) {

	url := "/0/connectivity/domain?configonly=true&level=3"
	body, _ := RestCall("GET", url, nil)
	var domain_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		panic(err)
	}
	for domains, domain_val := range domain_data {
		if nes, ok := domain_val.(map[string]interface{})["ne"]; ok {
			if _, ok := nes.(map[string]interface{})["bri"+ID]; ok {
				domainid = domains
				break
			}
		}
	}
	return
}

func FindNetworkGateway(domainID string, networkID string) (gatewayIP string) {

	url := "/0/connectivity/domain_prop/" + domainID + "/ne/brii" + networkID
	body, _ := RestCall("GET", url, nil)
	var domain_prop map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_prop)
	if err != nil {
		panic(err)
	}
	gatewayIP = domain_prop["ne_metadata"].(string)
	return
}

func DomainCreate(domainID string) {

	url := "/0/connectivity/domain?configonly=true"
	body, _ := RestCall("GET", url, nil)
	var domain_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		panic(err)
	}
	for domains := range domain_data {
		if domains == domainID {
			return
		}
	}

	url = "/0/tenant_manager/tenants" + "/" + domainID
	data := []byte(`{"containers": {
				"` + domainID + `": {
					"enable": "True",
					"qos_marking": "9",
					"context": "` + domainID + `",
					"type": "Gold",
					"property": "Container ` + domainID + ` Property",
					"domains": {}, "rules": {}}}}`)

	RestCall("PUT", url, data)

	url = "/0/tunnel_service/vnd_config/" + domainID
	data = []byte(`{"profile_name": "VXLAN",
			"add_vlan": "False"}`)
	RestCall("PUT", url, data)

	url = "/0/connectivity/domain/" + domainID
	data = []byte(`{"container_group": "` + domainID + `",
			"topology_name": "docker"}`)
	RestCall("PUT", url, data)

	url = "/0/connectivity/domain_prop/" + domainID
	data = []byte(`{}`)
	RestCall("PUT", url, data)

	url = "/0/pem_master/log_rule/" + domainID
	data = []byte(`{"rule": {
				"rule_` + domainID + `": {
					"pgtag1": "` + domainID + `",
					"log_ifc_type": "ACCESS_VM"}}}`)

	RestCall("PUT", url, data)
}

func DomainDelete(domainID string) {

	url := "/0/connectivity/domain?configonly=true"
	body, _ := RestCall("GET", url, nil)
	var domain_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		panic(err)
	}
	for domains, domain_val := range domain_data {
		if domains == domainID {
			res := domain_val.(map[string]interface{})["ne"]
			if len(res.(map[string]interface{})) == 0 {

				url = "/0/pem_master/log_rule/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/connectivity/domain_prop/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/connectivity/domain/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/tunnel_service/vnd_config/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/tenant_manager/tenants" + "/" + domainID
				RestCall("DELETE", url, nil)
			}
		}
	}
}

func GetNeId(NeName string, DomainId string) (NeId string) {

        url := "/0/connectivity/domain/" + DomainId + "/ne?configonly=true&level=1"
        body, _ := RestCall("GET", url, nil)
        var ne_data map[string]interface{}
        err := json.Unmarshal([]byte(body), &ne_data)
        if err != nil {
                panic(err)
        }
        for nes, ne_val := range ne_data {
                if ne_val.(map[string]interface{})["ne_dname"] == NeName {
                    NeId = nes
                    break
                }
        }
        return
}

func CreateRouterInterface(RouterName string, DomainId string, NetworkId string, IP string, Netmask string) {

        RouterId := GetNeId(RouterName, DomainId)
        CheckNeChildList(RouterId, DomainId, "ifc")
        CheckNeChildList(NetworkName(NetworkId), DomainId, "ifc")
        rtr_ifc := NetworkId
        url := "/0/connectivity/domain/" + DomainId + "/ne/" + RouterId + "/ifc/" + rtr_ifc
        data := []byte(`{"attachable": "true",
                         "list": "true",
                         "attach_type": "static,dynamic",
                         "mobility": "true",
                         "ifc_name": "`+rtr_ifc+`",
                         "ifc_type": "static",
                         "ip_address": "`+IP+`",
                         "ip_address_mask": "`+Netmask+`"}`)
        RestCall("PUT", url, data)

        net_ifc := RouterId
        url = "/0/connectivity/domain/" + DomainId + "/ne/" + NetworkName(NetworkId) + "/ifc/" + net_ifc
        data = []byte(`{"attachable": "true",
                         "list": "true",
                         "attach_type": "static,dynamic",
                         "mobility": "true",
                         "ifc_type": "static"}`)
        RestCall("PUT", url, data)

        link_name := RouterId + NetworkId
        url = "/0/connectivity/domain/" + DomainId + "/link/" + link_name
        data = []byte(`{"link_type": "static",
                         "link_name": "`+link_name+`",
                         "attachment1": "/ne/` + RouterId + `/ifc/` + rtr_ifc + `",
                         "attachment2": "/ne/` + NetworkName(NetworkId) + `/ifc/` + net_ifc + `"}`)
        RestCall("PUT", url, data)
}

func DeleteRouterInterface(RouterName string, DomainID string, NetworkID string) {

}

func NetworkName(ID string) (name string) {
        name = "bri" + ID
        return
}

func CheckNeChildList(NeId string, DomainId string, childList string) {

        url := "/0/connectivity/domain/" +  DomainId + "/ne/" + NeId + "?configonly=true"
        body, _ := RestCall("GET", url, nil)
        var ne_data map[string]interface{}
        err := json.Unmarshal([]byte(body), &ne_data)
        if err != nil {
                panic(err)
        }
        if _, ok := ne_data[childList]; ok {
                return
        } else {
                data := []byte(`{}`)
                RestCall("PUT", url + "/" + childList, data)
                return
        }
}
