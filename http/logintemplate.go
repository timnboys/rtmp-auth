package http

import (
	"html/template"

	"github.com/timnboys/rtmp-auth/storage"
)

type LoginPageTemplateData struct {
	State        *storage.State
	Config       ServerConfig
	CsrfTemplate template.HTML
	Errors       []error
}

var logintemplates = template.Must(template.New("logon.html").Parse(
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
<div class="login-container">
        <h2>Login</h2>
        <form action="/login" method="POST">
            <label for="username">Username:</label>
            <input type="text" id="username" name="username" required><br>
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required><br>
            <button type="submit">Login</button>
        </form>
    </div>
<script src="{{.Config.Prefix}}/public/main.js"></script>
</body>
</html>`))
