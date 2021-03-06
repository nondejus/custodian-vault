package btc

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/vault/logical"
)

func TestAddress(t *testing.T) {
	b, storage := getTestBackend(t)

	name := "test"
	network := "testnet"
	_, err := newWallet(t, b, storage, name, network, !segwitCompatible)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := newAuthToken(t, b, storage, name)
	if err != nil {
		t.Fatal(err)
	}
	token := resp.Data["token"].(string)

	t.Run("Get address for wallet", func(t *testing.T) {
		resp, err := newAddress(t, b, storage, name, token)
		if err != nil {
			t.Fatal(err)
		}
		if resp == nil {
			t.Fatal("No response received")
		}

		address := resp.Data["address"].(string)
		if !strings.HasPrefix(address, "m") && !strings.HasPrefix(address, "n") {
			t.Fatal("Invalid address:", address)
		}
		t.Log("Address:", address)
	})

	t.Run("Get address for BIP49 wallet", func(t *testing.T) {
		name := "segwit"
		_, err := newWallet(t, b, storage, name, network, segwitCompatible)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := newAuthToken(t, b, storage, name)
		if err != nil {
			t.Fatal(err)
		}
		token := resp.Data["token"].(string)

		resp, err = newAddress(t, b, storage, name, token)
		if err != nil {
			t.Fatal(err)
		}

		if resp == nil {
			t.Fatal("No response received")
		}

		address := resp.Data["address"].(string)
		if !strings.HasPrefix(address, "2") {
			t.Fatal("Invalid address:", address)
		}

		t.Log("Address:", address)
	})

	t.Run("Get address with expired auth token should fail", func(t *testing.T) {
		t.Parallel()

		exp := InvalidTokenError
		_, err := newAddress(t, b, storage, name, token)
		if err == nil {
			t.Fatal("Should have failed before")
		}
		if err.Error() != exp {
			t.Fatalf("Want: %v, got %v", exp, err)
		}
	})

	t.Run("Get address without auth token should fail", func(t *testing.T) {
		t.Parallel()

		token := ""
		exp := MissingTokenError
		_, err := newAddress(t, b, storage, name, token)
		if err == nil {
			t.Fatal("Should have failed before")
		}
		if err.Error() != exp {
			t.Fatalf("Want: %v, got: %v", exp, err)
		}
	})

	t.Run("Get address with invalid auth token should fail", func(t *testing.T) {
		t.Parallel()

		token := "testtoken"
		exp := InvalidTokenError
		_, err := newAddress(t, b, storage, name, token)
		if err == nil {
			t.Fatal("Should have failed before")
		}
		if err.Error() != exp {
			t.Fatalf("Want: %v, got: %v", exp, err)
		}
	})
}

func newAddress(t *testing.T, b logical.Backend, store logical.Storage, name string, token string) (*logical.Response, error) {
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Storage:   store,
		Path:      "address/" + name,
		Operation: logical.UpdateOperation,
		Data:      map[string]interface{}{"token": token},
	})
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error()
	}

	return resp, nil
}
