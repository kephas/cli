package sshkeys

import "testing"

func TestIsEncrypted(t *testing.T) {
	pk := &PrivateKey{Path: "n/a", Block: unencryptedPEM}
	if pk.IsEncrypted() {
		t.Error("expected false")
	}

	pk = &PrivateKey{Path: "n/a", Block: aesRSAPEM}
	if !pk.IsEncrypted() {
		t.Error("expected true")
	}

	pk = &PrivateKey{Path: "n/a", Block: des3RSAPEM}
	if !pk.IsEncrypted() {
		t.Error("expected true")
	}
}

func TestIsCipherImplemented(t *testing.T) {
	pk := &PrivateKey{Path: "n/a", Block: aesRSAPEM}
	if !pk.IsCipherImplemented(pk.cipher()) {
		t.Error("expected true")
	}

	pk = &PrivateKey{Path: "n/a", Block: des3RSAPEM}
	if !pk.IsCipherImplemented(pk.cipher()) {
		t.Error("expected true")
	}

	if pk.IsCipherImplemented("ANOTHER_CIPHER") {
		t.Error("expected false")
	}
}
