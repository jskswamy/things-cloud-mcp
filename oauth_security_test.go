package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestOAuthCredentialsAreEncryptedAtRestAndReloaded(t *testing.T) {
	t.Setenv("CREDENTIALS_SECRET", "")
	dir := t.TempDir()
	o := NewOAuthServer(NewUserManager(), dir)
	const email = "person@example.com"
	const password = "dummy-test-password"
	if err := o.persistCredential(email, password); err != nil {
		t.Fatalf("persist credential: %v", err)
	}
	refresh := &RefreshToken{
		Token: "refresh-test", Email: email, Password: password,
		ClientID: "client-test", ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := o.persistRefreshToken(refresh); err != nil {
		t.Fatalf("persist refresh token: %v", err)
	}

	var storedCredential, storedRefresh string
	if err := o.db.QueryRow(`SELECT password FROM credentials WHERE email=?`, email).Scan(&storedCredential); err != nil {
		t.Fatalf("read stored credential: %v", err)
	}
	if err := o.db.QueryRow(`SELECT password FROM refresh_tokens WHERE token=?`, refreshTokenKey(refresh.Token)).Scan(&storedRefresh); err != nil {
		t.Fatalf("read stored refresh credential: %v", err)
	}
	for name, stored := range map[string]string{"credential": storedCredential, "refresh": storedRefresh} {
		if stored == password || !strings.HasPrefix(stored, encryptedPasswordPrefix) {
			t.Fatalf("%s was not encrypted at rest", name)
		}
	}
	if err := o.db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	reloaded := NewOAuthServer(NewUserManager(), dir)
	defer reloaded.db.Close()
	if got := reloaded.credentials[email]; got != password {
		t.Fatalf("reloaded credential mismatch")
	}
	if got := reloaded.refreshTokens[refreshTokenKey(refresh.Token)]; got == nil || got.Password != password {
		t.Fatalf("reloaded refresh credential mismatch")
	}
	var storedToken string
	if err := reloaded.db.QueryRow(`SELECT token FROM refresh_tokens LIMIT 1`).Scan(&storedToken); err != nil {
		t.Fatalf("read stored refresh token: %v", err)
	}
	if storedToken == refresh.Token || !strings.HasPrefix(storedToken, refreshTokenHashPrefix) {
		t.Fatal("refresh token was not hashed at rest")
	}

	for _, path := range []string{dir, dir + "/oauth.db", dir + "/credentials.key"} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		want := os.FileMode(0600)
		if info.IsDir() {
			want = 0700
		}
		if got := info.Mode().Perm(); got != want {
			t.Fatalf("permissions for %s = %o, want %o", path, got, want)
		}
	}
}

func TestOAuthMigratesLegacyPlaintextCredentials(t *testing.T) {
	t.Setenv("CREDENTIALS_SECRET", "")
	dir := t.TempDir()
	o := NewOAuthServer(NewUserManager(), dir)
	const email = "legacy@example.com"
	const password = "legacy-dummy-password"
	if _, err := o.db.Exec(`INSERT INTO credentials(email, password) VALUES(?, ?)`, email, password); err != nil {
		t.Fatalf("insert legacy credential: %v", err)
	}
	const rawRefresh = "legacy-raw-refresh-token"
	if _, err := o.db.Exec(`INSERT INTO refresh_tokens VALUES(?,?,?,?,?)`, rawRefresh, email, password, "legacy-client", time.Now().Add(time.Hour).Format(time.RFC3339)); err != nil {
		t.Fatalf("insert legacy refresh token: %v", err)
	}
	if err := o.db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	migrated := NewOAuthServer(NewUserManager(), dir)
	defer migrated.db.Close()
	if got := migrated.credentials[email]; got != password {
		t.Fatalf("legacy credential was not loaded")
	}
	var stored string
	if err := migrated.db.QueryRow(`SELECT password FROM credentials WHERE email=?`, email).Scan(&stored); err != nil {
		t.Fatalf("read migrated credential: %v", err)
	}
	if stored == password || !strings.HasPrefix(stored, encryptedPasswordPrefix) {
		t.Fatal("legacy credential remained plaintext")
	}
	var storedToken, storedRefreshPassword string
	if err := migrated.db.QueryRow(`SELECT token, password FROM refresh_tokens LIMIT 1`).Scan(&storedToken, &storedRefreshPassword); err != nil {
		t.Fatalf("read migrated refresh token: %v", err)
	}
	if storedToken != refreshTokenKey(rawRefresh) || storedRefreshPassword == password || !strings.HasPrefix(storedRefreshPassword, encryptedPasswordPrefix) {
		t.Fatal("legacy refresh token or credential was not migrated")
	}
	if got := migrated.refreshTokens[refreshTokenKey(rawRefresh)]; got == nil || got.Password != password {
		t.Fatal("migrated refresh token was not loaded")
	}
}
