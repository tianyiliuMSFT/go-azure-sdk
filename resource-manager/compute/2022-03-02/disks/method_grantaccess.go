package disks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-sdk/sdk/client"
	"github.com/hashicorp/go-azure-sdk/sdk/client/pollers"
	"github.com/hashicorp/go-azure-sdk/sdk/client/resourcemanager"
	"github.com/hashicorp/go-azure-sdk/sdk/odata"
)

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type GrantAccessOperationResponse struct {
	Poller       pollers.Poller
	HttpResponse *http.Response
	OData        *odata.OData
	Model        *AccessUri
}

// GrantAccess ...
func (c DisksClient) GrantAccess(ctx context.Context, id commonids.ManagedDiskId, input GrantAccessData) (result GrantAccessOperationResponse, err error) {
	opts := client.RequestOptions{
		ContentType: "application/json; charset=utf-8",
		ExpectedStatusCodes: []int{
			http.StatusAccepted,
			http.StatusOK,
		},
		HttpMethod: http.MethodPost,
		Path:       fmt.Sprintf("%s/beginGetAccess", id.ID()),
	}

	req, err := c.Client.NewRequest(ctx, opts)
	if err != nil {
		return
	}

	if err = req.Marshal(input); err != nil {
		return
	}

	var resp *client.Response
	resp, err = req.Execute(ctx)
	if resp != nil {
		result.OData = resp.OData
		result.HttpResponse = resp.Response
		result.Model = &AccessUri{}
	}
	if err != nil {
		return
	}

	result.Poller, err = resourcemanager.PollerFromResponse(resp, c.Client)
	if err != nil {
		return
	}

	return
}

// GrantAccessThenPoll performs GrantAccess then polls until it's completed
func (c DisksClient) GrantAccessThenPoll(ctx context.Context, id commonids.ManagedDiskId, input GrantAccessData) (uri AccessUri, err error) {
	result, err := c.GrantAccess(ctx, id, input)
	if err != nil {
		return uri, fmt.Errorf("performing GrantAccess: %+v", err)
	}

	if err := result.Poller.PollUntilDone(ctx); err != nil {
		return uri, fmt.Errorf("polling after GrantAccess: %+v", err)
	}

	if err := result.Poller.FinalResult(result.Model); err != nil {
		return uri, fmt.Errorf("performing FinalResult: %+v", err)
	}

	bodyBytes, err := io.ReadAll(result.HttpResponse.Body)
	if err != nil {
		log.Fatalf("read body failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(result.HttpResponse.Body)

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, bodyBytes, "", "  ")
	if err != nil {
		log.Fatalf("invalid JSON: %v", err)
	}

	fmt.Println("Response body (pretty):")
	fmt.Println(prettyJSON.String())

	uri = *result.Model

	return
}
