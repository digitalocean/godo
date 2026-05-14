# Godo

[![GitHub Actions CI](https://github.com/digitalocean/godo/actions/workflows/ci.yml/badge.svg)](https://github.com/digitalocean/godo/actions/workflows/ci.yml)
[![Security](https://github.com/digitalocean/godo/actions/workflows/security.yml/badge.svg)](https://github.com/digitalocean/godo/actions/workflows/security.yml)
[![GoDoc](https://godoc.org/github.com/digitalocean/godo?status.svg)](https://godoc.org/github.com/digitalocean/godo)
[![Go Report Card](https://goreportcard.com/badge/github.com/digitalocean/godo)](https://goreportcard.com/report/github.com/digitalocean/godo)
[![codecov](https://codecov.io/gh/digitalocean/godo/branch/main/graph/badge.svg)](https://codecov.io/gh/digitalocean/godo)

Godo is a Go client library for accessing the DigitalOcean V2 API.

## 📚 Documentation

- **Client API docs**: [http://godoc.org/github.com/digitalocean/godo](http://godoc.org/github.com/digitalocean/godo)
- **DigitalOcean API docs**: [https://docs.digitalocean.com/reference/api/api-reference/](https://docs.digitalocean.com/reference/api/api-reference/)

## 🚀 Installation

```sh
go get github.com/digitalocean/godo@vX.Y.Z
```

where X.Y.Z is the [version](https://github.com/digitalocean/godo/releases) you need.

For the latest version:
```sh
go get github.com/digitalocean/godo
```

## 📖 Usage

```go
import "github.com/digitalocean/godo"
```

Create a new DigitalOcean client, then use the exposed services to access different parts of the DigitalOcean API.

### 🔐 Authentication

Currently, Personal Access Token (PAT) is the only method of authenticating with the API. You can manage your tokens at the DigitalOcean Control Panel [Applications Page](https://cloud.digitalocean.com/settings/applications).

```go
package main

import (
    "github.com/digitalocean/godo"
)

func main() {
    client := godo.NewFromToken("my-digitalocean-api-token")
}
```

If you need to provide a `context.Context` to your new client, use [`godo.NewClient`](https://godoc.org/github.com/digitalocean/godo#NewClient) to manually construct a client instead.

## 💡 Examples

### Creating a Droplet

```go
dropletName := "super-cool-droplet"

createRequest := &godo.DropletCreateRequest{
    Name:   dropletName,
    Region: "nyc3",
    Size:   "s-1vcpu-1gb",
    Image: godo.DropletCreateImage{
        Slug: "ubuntu-20-04-x64",
    },
}

ctx := context.TODO()

newDroplet, _, err := client.Droplets.Create(ctx, createRequest)
if err != nil {
    fmt.Printf("Something bad happened: %s\n\n", err)
    return err
}
```

### 📄 Pagination

If a list of items is paginated by the API, you must request pages individually. For example, to fetch all Droplets:

```go
func DropletList(ctx context.Context, client *godo.Client) ([]godo.Droplet, error) {
    // create a list to hold our droplets
    list := []godo.Droplet{}

    // create options. initially, these will be blank
    opt := &godo.ListOptions{}
    for {
        droplets, resp, err := client.Droplets.List(ctx, opt)
        if err != nil {
            return nil, err
        }

        // append the current page's droplets to our list
        list = append(list, droplets...)

        // if we are at the last page, break out the for loop
        if resp.Links == nil || resp.Links.IsLastPage() {
            break
        }

        page, err := resp.Links.CurrentPage()
        if err != nil {
            return nil, err
        }

        // set the page we want for the next request
        opt.Page = page + 1
    }

    return list, nil
}
```

### Token-based Pagination

Some endpoints offer token-based pagination. For example, to fetch all Registry Repositories:

```go
func ListRepositoriesV2(ctx context.Context, client *godo.Client, registryName string) ([]*godo.RepositoryV2, error) {
    list := []*godo.RepositoryV2{}
    opt := &godo.TokenListOptions{}
    
    for {
        repositories, resp, err := client.Registry.ListRepositoriesV2(ctx, registryName, opt)
        if err != nil {
            return nil, err
        }

        list = append(list, repositories...)

        if resp.Links == nil || resp.Links.IsLastPage() {
            break
        }

        nextPageToken, err := resp.Links.NextPageToken()
        if err != nil {
            return nil, err
        }

        opt.Token = nextPageToken
    }

    return list, nil
}
```

### 🔄 Automatic Retries and Exponential Backoff

The Godo client can be configured to use automatic retries and exponential backoff for requests that fail with 429 or 500-level response codes via [go-retryablehttp](https://github.com/hashicorp/go-retryablehttp). To configure Godo to enable usage of go-retryablehttp, the `RetryConfig.RetryMax` must be set.

```go
tokenSrc := oauth2.StaticTokenSource(&oauth2.Token{
    AccessToken: "dop_v1_xxxxxx",
})

oauth_client := oauth2.NewClient(oauth2.NoContext, tokenSrc)

waitMax := godo.PtrTo(6.0)
waitMin := godo.PtrTo(3.0)

retryConfig := godo.RetryConfig{
    RetryMax:     3,
    RetryWaitMin: waitMin,
    RetryWaitMax: waitMax,
}

client, err := godo.New(oauth_client, godo.WithRetryAndBackoffs(retryConfig))
```

Please refer to the [RetryConfig Godo documentation](https://pkg.go.dev/github.com/digitalocean/godo#RetryConfig) for more information.

## 🏷️ Versioning

Each version of the client is tagged and the version is updated accordingly. To see the list of past versions, run `git tag`.

## 🤝 Contributing

We love pull requests! Please see the [contribution guidelines](CONTRIBUTING.md).

## 📄 License

This library is distributed under the Apache 2.0 license found in the [LICENSE](./LICENSE.txt) file.
