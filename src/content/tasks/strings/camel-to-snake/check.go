package main

func TestToSnakeSimple(t *testing.T) {
	cases := map[string]string{
		"userName":  "user_name",
		"UserName":  "user_name",
		"id":        "id",
		"already_x": "already_x",
	}
	for in, want := range cases {
		if got := ToSnake(in); got != want {
			t.Fatalf("ToSnake(%q) = %q, ожидали %q", in, got, want)
		}
	}
}

func TestToSnakeAcronyms(t *testing.T) {
	cases := map[string]string{
		"parseURL":        "parse_url",
		"HTTPServer":      "http_server",
		"userID":          "user_id",
		"ID":              "id",
		"getHTTPResponse": "get_http_response",
	}
	for in, want := range cases {
		if got := ToSnake(in); got != want {
			t.Fatalf("ToSnake(%q) = %q, ожидали %q", in, got, want)
		}
	}
}

func TestToSnakeDigits(t *testing.T) {
	if got := ToSnake("base64Encode"); got != "base64_encode" {
		t.Fatalf(`ToSnake("base64Encode") = %q, ожидали "base64_encode"`, got)
	}
	if got := ToSnake(""); got != "" {
		t.Fatalf(`ToSnake("") = %q`, got)
	}
}
