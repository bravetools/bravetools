package platform

import "testing"

func TestParseImageString(t *testing.T) {
	name := "local:alpine/3.16/amd64"
	image, err := ParseImageString(name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%+v", image)
}

func TestParseImageStringDefaultArch(t *testing.T) {
	name := "local:alpine/3.16"
	image, err := ParseImageString(name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%+v", image)
}

func TestParseImageStringDefaultVersion(t *testing.T) {
	name := "local:alpine"
	image, err := ParseImageString(name)
	if err != nil {
		t.Fatalf("expected err when parsing image from filename %q", name)
	}
	t.Logf("%+v", image)
}

func TestParseImageStringNoName(t *testing.T) {
	name := "local:"
	_, err := ParseImageString(name)
	if err == nil {
		t.Fatalf("expected err when parsing image from filename %q", name)
	}
}

func TestParseImageStringLegacy(t *testing.T) {
	name := "alpine-3.16"
	image, err := ParseLegacyImageString(name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%+v", image)
}

func TestImageFromFilename(t *testing.T) {
	name := "alpine_3.16_amd64.tar.gz"
	image, err := ImageFromFilename(name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%+v", image)
}

func TestImageFromFilenameLong(t *testing.T) {
	name := "alpine_3.16_amd64_test.tar.gz"
	_, err := ImageFromFilename(name)
	if err == nil {
		t.Fatalf("expected err when parsing image from filename %q", name)
	}
}

func TestImageFromFilenameShort(t *testing.T) {
	name := "alpine_3.16_amd64_test.tar.gz"
	_, err := ImageFromFilename(name)
	if err == nil {
		t.Fatalf("expected err when parsing image from filename %q", name)
	}
}

func TestImageFromFilenameReal(t *testing.T) {
	name := "python-auth-1.0_1.0_amd64.tar.gz"
	image, err := ImageFromFilename(name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%+v", image)
}

func TestImageFromFilenameLegacy(t *testing.T) {
	name := "python-api-1.0.tar.gz"
	image, err := ImageFromFilename(name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%+v", image)
}

func TestImageFromFilenameIncorrectLong(t *testing.T) {
	name := "python-auth-1.0_1.0_amd64_TEST.tar.gz"
	_, err := ImageFromFilename(name)
	if err == nil {
		t.Fatalf("expected err when parsing image from filename %q", name)
	}
}
