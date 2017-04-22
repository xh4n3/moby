package client

import (
	"io"
	"net/url"

	"golang.org/x/net/context"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
)

// ImageCreate creates a new image based in the parent options.
// It returns the JSON content in the response body.
// 创建 Image，image 的 FULL URL 传给 parentReference
func (cli *Client) ImageCreate(ctx context.Context, parentReference string, options types.ImageCreateOptions) (io.ReadCloser, error) {
	ref, err := reference.ParseNormalizedNamed(parentReference)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	// 如果是默认的镜像，去除前缀
	query.Set("fromImage", reference.FamiliarName(ref))
	// 返回 Tag 或者 Digest
	query.Set("tag", getAPITagFromNamedRef(ref))
	// TODO: 目前只有发现这里是有 try 这个方法的，不知道为什么要这样做
	// server 端是有 router.WithCancel，逻辑貌似是在 Server 端实现
	resp, err := cli.tryImageCreate(ctx, query, options.RegistryAuth)
	if err != nil {
		return nil, err
	}
	return resp.body, nil
}

func (cli *Client) tryImageCreate(ctx context.Context, query url.Values, registryAuth string) (serverResponse, error) {
	headers := map[string][]string{"X-Registry-Auth": {registryAuth}}
	return cli.post(ctx, "/images/create", query, nil, headers)
}
