package templates

import (
  "strings"
)

func Layout(content string) string {
	html := `
<!DOCTYPE html>
<html ng-app="mailhogApp">
  <head>
    <title>MailHog</title>
    <script src="//code.jquery.com/jquery-1.11.0.min.js"></script>
    <link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css">
    <link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap-theme.min.css">
    <script src="//netdna.bootstrapcdn.com/bootstrap/3.1.1/js/bootstrap.min.js"></script>
    <script src="//cdnjs.cloudflare.com/ajax/libs/angular.js/1.2.15/angular.js"></script>
    <script src="/js/controllers.js"></script>
    <style>
      body, html { height: 100%; overflow: none; }
      .navbar {
        margin-bottom: 0;
        position: absolute;
        top: 0;
        right: 0;
        width: 100%;
      }
      .messages {
        padding-top: 50px;
      }
      .navbar-header img {
        height: 35px;
        margin: 8px 0 0 5px;
        float: left;
      }
      .navbar-nav.navbar-right:last-child {
        margin-right: 0; /* bootstrap fix?! */
      }
    </style>
  </head>
  <body ng-controller="MailCtrl">
    <nav class="navbar navbar-default navbar-static-top" role="navigation">
      <div class="navbar-header">
        <img src="/images/hog.png">
        <a class="navbar-brand" href="#">MailHog</a>
      </div>
      <ul class="nav navbar-nav navbar-right">
        <li class="dropdown">
          <a href="#" class="dropdown-toggle" data-toggle="dropdown">Options <b class="caret"></b></a>
          <ul class="dropdown-menu">
            <li><a href="#" ng-click="refresh()">Refresh</a></li>
            <li class="divider"></li>
            <li><a href="#" ng-click="deleteAll()">Delete all messages</a></li>
          </ul>
        </li>
        <li><a target="_blank" href="https://github.com/ian-kent/Go-MailHog">GitHub</a></li>
      </ul>
    </nav>
    <%= content %>
  </body>
</html>
`;
  return strings.Replace(html, "<%= content %>", content, -1);
}