package scanner

import "testing"

func TestScanFindsAndRedactsKnownSecret(t *testing.T) {
	secret := "AKIA" + "ABCDEFGHIJKLMNOP"
	got := mustScan(t, []Candidate{{File: "config.go", Line: 7, Text: `key := "` + secret + `"`}})
	if len(got) != 1 || got[0].Rule != "aws-access-token" {
		t.Fatalf("got %#v", got)
	}
}

func TestKnownSecretDoesNotAlsoProduceGenericFinding(t *testing.T) {
	secret := "sk_" + "live_ABCDEFGHIJKLMNOPQRST"
	got := mustScan(t, []Candidate{{File: ".env", Line: 1, Text: "STRIPE_SECRET_KEY=" + secret}})
	if len(got) != 1 || got[0].Rule != "stripe-access-token" {
		t.Fatalf("got %#v", got)
	}
}

func TestScanAllowsPlaceholdersAndExplicitSuppressions(t *testing.T) {
	got := mustScan(t, []Candidate{
		{File: ".env.example", Line: 1, Text: "API_KEY=your_api_key"},
		{File: "test.go", Line: 2, Text: "TOKEN=aVeryLongRandomishValue123 // dotenvai:allow test fixture"},
	})
	if len(got) != 0 {
		t.Fatalf("got %#v", got)
	}
}

func TestGitleaksProviderLibraryExpandsCoverage(t *testing.T) {
	secret := "glpat-" + "abcdefghijklmnopqrst"
	got := mustScan(t, []Candidate{{File: ".gitlab-ci.yml", Line: 12, Text: "GITLAB_TOKEN=" + secret}})
	if len(got) != 1 || got[0].Rule != "gitlab-pat" {
		t.Fatalf("got %#v", got)
	}
}

func TestDotenvLayerFindsCredentialURL(t *testing.T) {
	password := "correct-horse-battery-staple"
	got := mustScan(t, []Candidate{{File: ".env", Line: 4, Text: "DATABASE_URL=postgres://app:" + password + "@db.internal/app"}})
	if len(got) != 1 || got[0].Rule != "credential-url" {
		t.Fatalf("got %#v", got)
	}
}

func mustScan(t *testing.T, candidates []Candidate) []Finding {
	t.Helper()
	findings, err := Scan(candidates)
	if err != nil {
		t.Fatal(err)
	}
	return findings
}

func TestParseUnifiedDiffTracksAddedLines(t *testing.T) {
	diff := "diff --git a/a b/a\n+++ b/a\n@@ -3,2 +3,3 @@\n same\n+TOKEN=abcDEF1234567890\n tail\n" // dotenvai:allow test fixture
	got := ParseUnifiedDiff(diff)
	if len(got) != 1 || got[0].File != "a" || got[0].Line != 4 {
		t.Fatalf("got %#v", got)
	}
}
