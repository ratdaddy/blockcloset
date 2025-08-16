package config

import "testing"

func TestAppEnvDefaultsToTestInGoTest(t *testing.T) {
	t.Setenv("APP_ENV", "")

	Init()

	if AppEnv != "test" {
		t.Fatalf("expected default AppEnv to be \"test\" when running tests, got %q", AppEnv)
	}
}
