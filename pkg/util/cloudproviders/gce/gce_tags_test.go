// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package gce

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configmock "github.com/DataDog/datadog-agent/pkg/config/mock"
	"github.com/DataDog/datadog-agent/pkg/util/cache"
)

var (
	expectedFullTags = []string{
		"tag",
		"zone:us-east1-b",
		"instance-type:n1-standard-1",
		"internal-hostname:dd-test.c.datadog-dd-test.internal",
		"instance-id:1111111111111111111",
		"project:test-project",
		"numeric_project_id:111111111111",
		"cluster-location:us-east1-b",
		"cluster-name:test-cluster",
		"created-by:projects/111111111111/zones/us-east1-b/instanceGroupManagers/gke-test-cluster-default-pool-0012834b-grp",
		"gci-ensure-gke-docker:true",
		"gci-update-strategy:update_disabled",
		"google-compute-enable-pcid:true",
		"instance-template:projects/111111111111/global/instanceTemplates/gke-test-cluster-default-pool-0012834b",
	}
	expectedTagsWithProjectID = append(expectedFullTags, "project_id:test-project")
	expectedExcludedTags      = []string{
		"tag",
		"zone:us-east1-b",
		"instance-type:n1-standard-1",
		"internal-hostname:dd-test.c.datadog-dd-test.internal",
		"instance-id:1111111111111111111",
		"project:test-project",
		"numeric_project_id:111111111111",
		"cluster-location:us-east1-b",
		"created-by:projects/111111111111/zones/us-east1-b/instanceGroupManagers/gke-test-cluster-default-pool-0012834b-grp",
		"gci-ensure-gke-docker:true",
		"gci-update-strategy:update_disabled",
		"google-compute-enable-pcid:true",
		"instance-template:projects/111111111111/global/instanceTemplates/gke-test-cluster-default-pool-0012834b",
	}
	expectedTagsWithProviderKind = append(expectedFullTags, "provider_kind:test-provider")
)

func mockMetadataRequest(t *testing.T) *httptest.Server {
	content, err := os.ReadFile("test/gce_metadata.json")
	if err != nil {
		assert.Fail(t, fmt.Sprintf("Error getting test data: %v", err))
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "/?recursive=true")
		assert.Equal(t, "Google", r.Header.Get("Metadata-Flavor"))
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	}))
	metadataURL = ts.URL
	return ts
}

func mockMetadataRequestError(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "/?recursive=true")
		assert.Equal(t, "Google", r.Header.Get("Metadata-Flavor"))
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "some GCE error", http.StatusInternalServerError)
	}))
	metadataURL = ts.URL
	return ts
}

func testTags(t *testing.T, tags []string, expectedTags []string) {
	require.Len(t, tags, len(expectedTags))
	for _, tag := range tags {
		assert.Contains(t, expectedTags, tag)
	}
}

func TestGetHostTags(t *testing.T) {
	ctx := context.Background()
	server := mockMetadataRequest(t)
	defer server.Close()
	defer cache.Cache.Delete(tagsCacheKey)
	tags, err := GetTags(ctx)
	require.NoError(t, err)
	testTags(t, tags, expectedFullTags)
}

func TestGetHostTagsWithProjectID(t *testing.T) {
	mockConfig := configmock.New(t)
	ctx := context.Background()
	server := mockMetadataRequest(t)
	defer server.Close()
	defer cache.Cache.Delete(tagsCacheKey)
	mockConfig.SetWithoutSource("gce_send_project_id_tag", true)
	tags, err := GetTags(ctx)
	require.NoError(t, err)
	testTags(t, tags, expectedTagsWithProjectID)
}

func TestGetHostTagsSuccessThenError(t *testing.T) {
	ctx := context.Background()
	server := mockMetadataRequest(t)
	tags, err := GetTags(ctx)
	require.NotNil(t, tags)
	require.NoError(t, err)
	server.Close()

	server = mockMetadataRequestError(t)
	defer server.Close()
	defer cache.Cache.Delete(tagsCacheKey)
	tags, err = GetTags(ctx)
	require.NoError(t, err)
	testTags(t, tags, expectedFullTags)
}

func TestGetHostTagsWithNonDefaultTagFilters(t *testing.T) {
	ctx := context.Background()
	mockConfig := configmock.New(t)
	defaultExclude := mockConfig.GetStringSlice("exclude_gce_tags")

	mockConfig.SetWithoutSource("exclude_gce_tags", append([]string{"cluster-name"}, defaultExclude...))

	server := mockMetadataRequest(t)
	defer server.Close()
	defer cache.Cache.Delete(tagsCacheKey)

	tags, err := GetTags(ctx)
	require.NoError(t, err)
	testTags(t, tags, expectedExcludedTags)
}

func TestGetHostTagsWithProviderKind(t *testing.T) {
	ctx := context.Background()
	mockConfig := configmock.New(t)
	mockConfig.SetWithoutSource("provider_kind", "test-provider")

	server := mockMetadataRequest(t)
	defer server.Close()
	defer cache.Cache.Delete(tagsCacheKey)

	tags, err := GetTags(ctx)
	require.NoError(t, err)
	testTags(t, tags, expectedTagsWithProviderKind)
}
