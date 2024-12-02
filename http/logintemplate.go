package http

import (
	"html/template"

	//"github.com/timnboys/rtmp-auth/storage"
)

type TemplateData struct {
	State        *storage.State
	Config       ServerConfig
	CsrfTemplate template.HTML
	Errors       []error
}

var templates = template.Must(template.New("form.html").Parse(
	`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>RTMP Login</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css" href="{{.Config.Prefix}}/public/mini-dark.css">
  <link rel="stylesheet" type="text/css" href="{{.Config.Prefix}}/public/main.css">
</head>
<body>
  <div class="container">
    <h1><a href="{{$.Config.Prefix}}">rtmp-auth</a></h1>
    <p>
			Login with the following,
	</p>
	<ul>
			<li><a href="/login-gl">Google</a></li>
	</ul>
<script src="{{.Config.Prefix}}/public/main.js"></script>
</body>
</html>`))
