package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/couchbase/gocb/v2"
	// "github.com/couchbase/go-couchbase"
)

/*
	{
	  "value": "TrzdyE9qjetRWN0tisqUIo8YjL26Mhz5E-ZXHc0KeyvLu99JuXmd_ZiKyg8T6MJQHyiQAlpQd5XvsTmve_2M9e2qC05lEuN92NkNgG0p3pyYyuauR8PvHC1y6pO4c3IqnjhHcE0eh1bXljSd5pqGLgzGe0PpBJnunGwDOy8-4LEkn2XuILFPxwXgovWICaUvmvZPLSqk7I8hpNbmN0V89Q",
	  "accountType": "WECOM",
	  "expireIn": 7200,
	  "createdAt": "2021-10-27T12:32:02.502Z",
	  "updatedAt": "2021-12-28T06:42:34.263Z",
	  "_type": "Token"
	}
*/
// {
// 	"value": "52_6nVKARh5ByQnTEWvm4H9ru_ECmvO20AVSNSa-qNcNYzWBwzmEjr1xfgcJNXALjRz9RJRaKHbs9aQtFh8UJzgiltp04ks453eN98SwXIPIQyHkX8roIerMkMxvWUrhR6Szn5LLqO2OkL9TlznZMMcABAIVV",
// 	"accountType": "MP",
// 	"expireIn": 7200,
// 	"createdAt": "2021-06-29T08:22:33.249Z",
// 	"updatedAt": "2021-12-28T07:01:13.331Z",
// 	"_type": "Token"
//   }

func conncb(cb_addr string, user_name string, password string) (*gocb.Cluster, error) {
	cluster, err := gocb.Connect(cb_addr, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: user_name,
			Password: password,
		},
	})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func main() {

	const staging_cb_addr = "172.16.5.35:8091"
	const staging_cb_user = "xxxxxx"
	const staging_cb_pwd = "xxxxxx"

	const prod_cb_addr = "172.16.5.58:8091"
	const prod_cb_user = "xxxxxx"
	const prod_cb_pwd = "xxxxxx"

	const token_id_wecom = "accessToken:wwc9a5cb2719ab6225"
	const token_id_mp = "accessToken:wx88ed27696de5b9a4"

	type Doc struct {
		ExpireIn    int    `json:"expireIn"`
		Type        string `json:"_type"`
		Value       string `json:"value"`
		AccountType string `json:"accountType"`
		CreateAt    string `json:"createdAt"`
		UpdateAt    string `json:"updatedAt"`
	}

	var srcdocument_wecom_token Doc
	var srcdocument_mp_token Doc

	cluster, err := conncb(prod_cb_addr, prod_cb_user, prod_cb_pwd)
	if err != nil {
		log.Fatal(err)
	}
	bucket := cluster.Bucket("wechat-connector")
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection := bucket.DefaultCollection()

	result, err := collection.Get(token_id_wecom, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = result.Content(&srcdocument_wecom_token)
	if err != nil {
		log.Fatal(err)
	}
	prettyJSON, err := json.MarshalIndent(srcdocument_wecom_token, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json for wecom token", err)
	}
	fmt.Printf("%s\n", string(prettyJSON))

	result, err = collection.Get(token_id_mp, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = result.Content(&srcdocument_mp_token)
	if err != nil {
		log.Fatal(err)
	}
	prettyJSON, err = json.MarshalIndent(srcdocument_mp_token, "", "    ")
	if err != nil {
		log.Fatal("Failed to generate json for mp token", err)
	}
	fmt.Printf("%s\n", string(prettyJSON))

	// connect staging couchbase ,and update token from prod couchbase.
	cluster, err = conncb(staging_cb_addr, staging_cb_user, staging_cb_pwd)
	if err != nil {
		log.Fatal(err)
	}

	bucket = cluster.Bucket("wechat-connector")
	err = bucket.WaitUntilReady(5*time.Second, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = bucket.DefaultCollection()
	_, err = collection.Upsert(token_id_wecom, &srcdocument_wecom_token, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("wecom token synced!")
	}
	collection.Upsert(token_id_mp, &srcdocument_mp_token, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("mp token synced!")
	}
}
