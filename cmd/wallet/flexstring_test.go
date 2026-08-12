package wallet_test

import (
	"encoding/json"
	"testing"

	"github.com/keeperhub/cli/cmd/wallet"
)

func TestFlexibleStringAcceptsBothShapes(t *testing.T) {
	// success cases: number and string both normalise to the same string value.
	for _, tc := range []struct{ in, want string }{
		{`{"chainId":11155111}`, "11155111"},   // current API — returned *json.UnmarshalTypeError before this fix
		{`{"chainId":"11155111"}`, "11155111"}, // legacy API — must keep working
		{`{"chainId":null}`, ""},               // null is a no-op; zero value preserved
	} {
		var cb wallet.ChainBalance
		if err := json.Unmarshal([]byte(tc.in), &cb); err != nil {
			t.Fatalf("ChainBalance %s: %v", tc.in, err)
		}
		if string(cb.ChainID) != tc.want {
			t.Fatalf("ChainBalance %s: got %q want %q", tc.in, cb.ChainID, tc.want)
		}

		var tok wallet.Token
		if err := json.Unmarshal([]byte(tc.in), &tok); err != nil {
			t.Fatalf("Token %s: %v", tc.in, err)
		}
		if string(tok.ChainID) != tc.want {
			t.Fatalf("Token %s: got %q want %q", tc.in, tok.ChainID, tc.want)
		}
	}

	// error cases: non-string, non-number JSON values must return an error.
	for _, bad := range []string{
		`{"chainId":{"x":1}}`,
		`{"chainId":true}`,
		`{"chainId":[1,2]}`,
	} {
		var cb wallet.ChainBalance
		if err := json.Unmarshal([]byte(bad), &cb); err == nil {
			t.Fatalf("ChainBalance %s: expected error, got nil", bad)
		}
		var tok wallet.Token
		if err := json.Unmarshal([]byte(bad), &tok); err == nil {
			t.Fatalf("Token %s: expected error, got nil", bad)
		}
	}
}

func TestFlexibleStringMarshalsNumericAsNumber(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{`{"chainId":11155111}`, `11155111`},   // numeric in -> numeric out
		{`{"chainId":"11155111"}`, `11155111`}, // legacy string in -> numeric out
		{`{"chainId":"mainnet"}`, `"mainnet"`}, // non-numeric stays a string
		{`{"chainId":null}`, `""`},             // null -> zero value
	} {
		var cb wallet.ChainBalance
		if err := json.Unmarshal([]byte(tc.in), &cb); err != nil {
			t.Fatalf("%s: %v", tc.in, err)
		}
		got, err := json.Marshal(cb.ChainID)
		if err != nil {
			t.Fatalf("%s: marshal: %v", tc.in, err)
		}
		if string(got) != tc.want {
			t.Fatalf("%s: got %s want %s", tc.in, got, tc.want)
		}
	}
}
