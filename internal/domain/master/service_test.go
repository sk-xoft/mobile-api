package master

import "testing"

func TestLocalizedUsesThaiByDefaultAndEnglishWhenRequested(t *testing.T) {
	value := "ภาษาไทย\x00English"
	if got := localized(value, false); got != "ภาษาไทย" {
		t.Fatalf("Thai localization = %q", got)
	}
	if got := localized(value, true); got != "English" {
		t.Fatalf("English localization = %q", got)
	}
}

func TestLocalizedFallsBackToThaiWhenEnglishIsMissing(t *testing.T) {
	if got := localized("ภาษาไทย", true); got != "ภาษาไทย" {
		t.Fatalf("fallback localization = %q", got)
	}
}
