var mailhogApp = angular.module('mailhogApp', []);

mailhogApp.directive('targetBlank', function(){
  return {
    link : function(scope, element, attributes){
      element.on('load', function() {
        var a = element.contents().find('a');
        a.attr('target', '_blank');
      });
    }
  };
});

function guid() {
  function s4() {
    return Math.floor((1 + Math.random()) * 0x10000)
               .toString(16)
               .substring(1);
  }
  return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
         s4() + '-' + s4() + s4() + s4();
}

mailhogApp.directive('ngKeyEnter', function () {
  return function (scope, element, attrs) {
    element.bind("keydown keypress", function (event) {
      if(event.which === 13) {
        scope.$apply(function (){
          scope.$eval(attrs.ngKeyEnter);
        });
        event.preventDefault();
      }
    });
  };
});

mailhogApp.controller('MailCtrl', function ($scope, $http, $sce, $timeout) {
  $scope.host = apiHost;

  $scope.cache = {};
  $scope.previewAllHeaders = false;

  $scope.eventsPending = {};
  $scope.eventCount = 0;
  $scope.eventDone = 0;
  $scope.eventFailed = 0;

  $scope.hasEventSource = false;
  $scope.source = null;

  $scope.itemsPerPage = 50
  $scope.startIndex = 0

  if(typeof(Storage) !== "undefined") {
      $scope.itemsPerPage = parseInt(localStorage.getItem("itemsPerPage"), 10)
      if(!$scope.itemsPerPage) {
        $scope.itemsPerPage = 50;
        localStorage.setItem("itemsPerPage", 50)
      }
  }

  $scope.startMessages = 0
  $scope.countMessages = 0
  $scope.totalMessages = 0

  $scope.startSearchMessages = 0
  $scope.countSearchMessages = 0
  $scope.totalSearchMessages = 0

  $scope.jim = null

  $scope.smtpmech = "NONE"
  $scope.selectedOutgoingSMTP = ""
  $scope.saveSMTPServer = false;

  $scope.getJim = function() {
    var url = $scope.host + 'api/v2/jim'
    $http.get(url).success(function(data) {
      $scope.jim = data
    }).error(function() {
      $scope.jim = null
    })
  }
  $scope.getJim()

  $scope.enableJim = function() {
    var url = $scope.host + 'api/v2/jim'
    $http.post(url).success(function(data) {
      $scope.getJim()
    })
  }
  $scope.disableJim = function() {
    var url = $scope.host + 'api/v2/jim'
    $http.delete(url).success(function(data) {
      $scope.getJim()
    })
  }

  $(function() {
    $scope.openStream();
    if(typeof(Notification) !== "undefined") {
      Notification.requestPermission();
    }
  });

  $scope.getMoment = function(a) {
    return moment(a)
  }

  $scope.backToInbox = function() {
    $scope.preview = null;
    $scope.searching = false;
  }
  $scope.backToInboxFirst = function() {
    $scope.preview = null;
    $scope.startIndex = 0;
    $scope.startMessages = 0;
    $scope.searching = false;
    $scope.refresh();
  }

  $scope.toggleStream = function() {
    $scope.source == null ? $scope.openStream() : $scope.closeStream();
  }
  $scope.openStream = function() {
    var host = $scope.host.replace(/^http/, 'ws') ||
               (location.protocol.replace(/^http/, 'ws') + '//' + location.hostname + (location.port ? ':' + location.port : '') + location.pathname);
    $scope.source = new WebSocket(host + 'api/v2/websocket');
    $scope.source.addEventListener('message', function(e) {
      $scope.$apply(function() {
        $scope.totalMessages++;
        if ($scope.startIndex > 0) {
          $scope.startIndex++;
          $scope.startMessages++;
          return
        }
        if ($scope.countMessages < $scope.itemsPerPage) {
          $scope.countMessages++;
        }
        var message = JSON.parse(e.data);
        $scope.messages.unshift(message);
        while($scope.messages.length > $scope.itemsPerPage) {
          $scope.messages.pop();
        }
        if(typeof(Notification) !== "undefined") {
          $scope.createNotification(message);
        }
      });
    }, false);
    $scope.source.addEventListener('open', function(e) {
      $scope.$apply(function() {
        $scope.hasEventSource = true;
      });
    }, false);
    $scope.source.addEventListener('error', function(e) {
      //if(e.readyState == EventSource.CLOSED) {
        $scope.$apply(function() {
          $scope.hasEventSource = false;
        });
      //}
    }, false);
  }
  $scope.closeStream = function() {
    $scope.source.close();
    $scope.source = null;
    $scope.hasEventSource = false;
  }

  $scope.createNotification = function(message) {
    var title = "Mail from " + $scope.getSender(message);
    var options = {
      body: $scope.tryDecodeMime(message.Content.Headers["Subject"][0]),
      tag: "MailHog",
      icon: "images/hog.png"
    };
    var notification = new Notification(title, options);
    notification.addEventListener('click', function(e) {
      $scope.selectMessage(message);
      window.focus();
      notification.close();
    });
  }

  $scope.tryDecodeMime = function(str) {
    return unescapeFromMime(str)
  }

  $scope.resizePreview = function() {
    $('.tab-content').height($(window).innerHeight() - $('.tab-content').offset().top);
    $('.tab-content .tab-pane').height($(window).innerHeight() - $('.tab-content').offset().top);
  }

  $scope.getSender = function(message) {
    return $scope.tryDecodeMime($scope.getDisplayName(message.Content.Headers["From"][0]) ||
                                message.From.Mailbox + "@" + message.From.Domain);
  }

  $scope.getDisplayName = function(value) {
    if(!value) { return ""; }

    res = value.match(/(.*)\<(.*)\>/);

    if(res) {
      if(res[1].trim().length > 0) {
        return res[1].trim();
      }
      return res[2];
    }
    return value
  }

  $scope.startEvent = function(name, args, glyphicon) {
    var eID = guid();
    //console.log("Starting event '" + name + "' with id '" + eID + "'")
    var e = {
      id: eID,
      name: name,
      started: new Date(),
      complete: false,
      failed: false,
      args: args,
      glyphicon: glyphicon,
      getClass: function() {
        // FIXME bit nasty
        if(this.failed) {
          return "bg-danger"
        }
        if(this.complete) {
          return "bg-success"
        }
        return "bg-warning"; // pending
      },
      done: function() {
        //delete $scope.eventsPending[eID]
        var e = this;
        e.complete = true;
        $scope.eventDone++;
        if(this.failed) {
          // console.log("Failed event '" + e.name + "' with id '" + eID + "'")
        } else {
          // console.log("Completed event '" + e.name + "' with id '" + eID + "'")
          $timeout(function() {
            e.remove();
          }, 10000);
        }
      },
      fail: function() {
        $scope.eventFailed++;
        this.failed = true;
        this.done();
      },
      remove: function() {
        // console.log("Deleted event '" + e.name + "' with id '" + eID + "'")
        if(e.failed) {
          $scope.eventFailed--;
        }
        delete $scope.eventsPending[eID];
        $scope.eventDone--;
        $scope.eventCount--;
        return false;
      }
    };
    $scope.eventsPending[eID] = e;
    $scope.eventCount++;
    return e;
  }

  $scope.messagesDisplayed = function() {
    return $('.messages .msglist-message').length
  }

  $scope.refresh = function() {
    if ($scope.searching) {
      return $scope.refreshSearch();
    }
    var e = $scope.startEvent("Loading messages", null, "glyphicon-download");
    var url = $scope.host + 'api/v2/messages'
    if($scope.startIndex > 0) {
      url += "?start=" + $scope.startIndex + "&limit=" + $scope.itemsPerPage;
    } else {
      url += "?limit=" + $scope.itemsPerPage;
    }
    $http.get(url).success(function(data) {
      $scope.messages = data.items;
      $scope.totalMessages = data.total;
      $scope.countMessages = data.count;
      $scope.startMessages = data.start;
      e.done();
    });
  }
  $scope.refresh();

  $scope.showNewer = function() {
    $scope.startIndex -= $scope.itemsPerPage;
    if($scope.startIndex < 0) {
      $scope.startIndex = 0
    }
    $scope.refresh();
  }

  $scope.showUpdated = function(i) {
    $scope.itemsPerPage = parseInt(i, 10);
    if(typeof(Storage) !== "undefined") {
        localStorage.setItem("itemsPerPage", $scope.itemsPerPage)
    }
    $scope.refresh();
  }

  $scope.showOlder = function() {
    $scope.startIndex += $scope.itemsPerPage;
    $scope.refresh();
  }

  $scope.search = function(kind, text) {
    $scope.searching = true;
    $scope.searchKind = kind;
    $scope.searchedText = text;
    $scope.searchText = "";
    $scope.startSearchMessages = 0
    $scope.countSearchMessages = 0
    $scope.totalSearchMessages = 0
    $scope.refreshSearch()
  }

  $scope.refreshSearch = function() {
    var url = $scope.host + 'api/v2/search?kind=' + $scope.searchKind + '&query=' + $scope.searchedText;
    if($scope.startIndex > 0) {
      url += "&start=" + $scope.startIndex;
    }
    $http.get(url).success(function(data) {
      $scope.searchMessages = data.items;
      $scope.totalSearchMessages = data.total;
      $scope.countSearchMessages = data.count;
      $scope.startSearchMessages = data.start;
    });
  }

  $scope.hasSelection = function() {
    return $(".messages :checked").length > 0 ? true : false;
  }

  $scope.selectMessage = function(message) {
    $timeout(function(){
      $scope.resizePreview();
    }, 0);
  	if($scope.cache[message.ID]) {
  		$scope.preview = $scope.cache[message.ID];
      //reflow();
  	} else {
  		$scope.preview = message;
      var e = $scope.startEvent("Loading message", message.ID, "glyphicon-download-alt");
	  	$http.get($scope.host + 'api/v1/messages/' + message.ID).success(function(data) {
	  	  $scope.cache[message.ID] = data;

        // FIXME
        // - nested mime parts can't be downloaded

        data.$cidMap = {};
        if(data.MIME && data.MIME.Parts.length) {
          for(p in data.MIME.Parts) {
            for(h in data.MIME.Parts[p].Headers) {
              if(h.toLowerCase() == "content-id") {
                cid = data.MIME.Parts[p].Headers[h][0]
                cid = cid.substr(1,cid.length-2)
                data.$cidMap[cid] = "api/v1/messages/" + message.ID + "/mime/part/" + p + "/download"
              }
            }
          }
        }
        console.log(data.$cidMap)
        // TODO
        // - scan HTML parts for elements containing CID URI and replace

        h = $scope.getMessageHTML(data)
        for(c in data.$cidMap) {
	  str = "cid:" + c;
	  pat = str.replace(/([.*+?^=!:${}()|\[\]\/\\])/g, "\\$1");
          h = h.replace(new RegExp(pat, 'g'), data.$cidMap[c])
        }
	      data.previewHTML = $sce.trustAsHtml(h);
  		  $scope.preview = data;
  		  preview = $scope.cache[message.ID];
        //reflow();
        e.done();
	    });
	   }
  }

  $scope.toggleHeaders = function(val) {
    $scope.previewAllHeaders = val;
    $timeout(function(){
      $scope.resizePreview();
    }, 0);
    var t = window.setInterval(function() {
      if(val) {
        if($('#hide-headers').length) {
          window.clearInterval(t);
          //reflow();
        }
      } else {
        if($('#show-headers').length) {
          window.clearInterval(t);
          //reflow();
        }
      }
    }, 10);
  }

  $scope.fileSize = function(bytes) {
    return filesize(bytes)
  }

  $scope.tryDecodeContent = function(message) {
    var charset = "UTF-8"
    if(message.Content.Headers["Content-Type"][0]) {
      // TODO
    }

    var content = message.Content.Body;
    var contentTransferEncoding = message.Content.Headers["Content-Transfer-Encoding"][0];

    if(contentTransferEncoding) {
      switch (contentTransferEncoding.toLowerCase()) {
        case 'quoted-printable':
          content = content.replace(/=[\r\n]+/gm,"");
          content = unescapeFromQuotedPrintableWithoutRFC2047(content, charset);
          break;
        case 'base64':
          // remove line endings to give original base64-encoded string
          content = content.replace(/\r?\n|\r/gm,"");
          content = unescapeFromBase64(content, charset);
          break;
      }
    }

    return content;
  }

  $scope.formatMessagePlain = function(message) {
    var body = $scope.getMessagePlain(message);
    var escaped = $scope.escapeHtml(body);
    var formatted = escaped.replace(/(https?:\/\/)([-[\]A-Za-z0-9._~:/?#@!$()*+,;=%]|&amp;|&#39;)+/g, '<a href="$&" target="_blank">$&</a>');
    return $sce.trustAsHtml(formatted);
  }

  $scope.escapeHtml = function(html) {
    var entityMap = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;'
    };
    return html.replace(/[&<>"']/g, function (s) {
      return entityMap[s];
    });
  }

  $scope.getMessagePlain = function(message) {
    if (message.Content.Headers && message.Content.Headers["Content-Type"] && message.Content.Headers["Content-Type"][0].match("text/plain")) {
      return $scope.tryDecode(message.Content);
    }
    var l = $scope.findMatchingMIME(message, "text/plain");
    if(l != null && l !== "undefined") {
      return $scope.tryDecode(l);
    }
    return message.Content.Body;
  }

  $scope.findMatchingMIME = function(part, mime) {
    // TODO cache results
    if(part.MIME) {
      for(var p in part.MIME.Parts) {
        if("Content-Type" in part.MIME.Parts[p].Headers) {
          if(part.MIME.Parts[p].Headers["Content-Type"].length > 0) {
            if(part.MIME.Parts[p].Headers["Content-Type"][0].match(mime + ";?.*")) {
              return part.MIME.Parts[p];
            } else if (part.MIME.Parts[p].Headers["Content-Type"][0].match(/multipart\/.*/)) {
              var f = $scope.findMatchingMIME(part.MIME.Parts[p], mime);
              if(f != null) {
                return f;
              }
            }
          }
        }
      }
    }
    return null;
  }
  $scope.hasHTML = function(message) {
    // TODO cache this
    for(var header in message.Content.Headers) {
      if(header.toLowerCase() == 'content-type') {
        if(message.Content.Headers[header][0].match("text/html")) {
          return true
        }
      }
    }

    var l = $scope.findMatchingMIME(message, "text/html");
    if(l != null && l !== "undefined") {
      return true
    }
    return false;
  }
  $scope.getMessageHTML = function(message) {
    console.log(message);
    for(var header in message.Content.Headers) {
      if(header.toLowerCase() == 'content-type') {
        if(message.Content.Headers[header][0].match("text/html")) {
          return $scope.tryDecode(message.Content);
        }
      }
    }

    var l = $scope.findMatchingMIME(message, "text/html");
    if(l != null && l !== "undefined") {
      return $scope.tryDecode(l);
    }
  	return "<HTML not found>";
	}

  $scope.tryDecode = function(l){
    if(l.Headers && l.Headers["Content-Type"] && l.Headers["Content-Transfer-Encoding"]){
      return $scope.tryDecodeContent({Content:l});
    }else{
      return l.Body;
    }
  };
  $scope.date = function(timestamp) {
  	return (new Date(timestamp)).toString();
  };

  $scope.deleteAll = function() {
  	$('#confirm-delete-all').modal('show');
  }

  $scope.releaseOne = function(message) {
    $scope.releasing = message;

    $http.get($scope.host + 'api/v2/outgoing-smtp').success(function(data) {
      $scope.outgoingSMTP = data;
      $('#release-one').modal('show');
    })
  }
  $scope.confirmReleaseMessage = function() {
    $('#release-one').modal('hide');
    var message = $scope.releasing;
    $scope.releasing = null;

    var e = $scope.startEvent("Releasing message", message.ID, "glyphicon-share");

    if($('#release-message-outgoing').val().length > 0) {
      authcfg = {
        name: $('#release-message-outgoing').val(),
        email: $('#release-message-email').val(),
      }
    } else {
      authcfg = {
        email: $('#release-message-email').val(),
        host: $('#release-message-smtp-host').val(),
        port: $('#release-message-smtp-port').val(),
        mechanism: $('#release-message-smtp-mechanism').val(),
        username: $('#release-message-smtp-username').val(),
        password: $('#release-message-smtp-password').val(),
        save: $('#release-message-save').is(":checked") ? true : false,
        name: $('#release-message-server-name').val(),
      }
    }

    $http.post($scope.host + 'api/v1/messages/' + message.ID + '/release', authcfg).success(function() {
      e.done();
    }).error(function(err) {
      e.fail();
      e.error = err;
    });
  }

  $scope.getSource = function(message) {
  	var source = "";
  	$.each(message.Content.Headers, function(k, v) {
  		source += k + ": " + v + "\n";
  	});
	source += "\n";
	source += message.Content.Body;
	return source;
  }

  $scope.deleteAllConfirm = function() {
  	$('#confirm-delete-all').modal('hide');
    var e = $scope.startEvent("Deleting all messages", null, "glyphicon-remove-circle");
  	$http.delete($scope.host + 'api/v1/messages').success(function() {
  		$scope.refresh();
  		$scope.preview = null;
      e.done()
  	});
  }

  $scope.deleteOne = function(message) {
    var e = $scope.startEvent("Deleting message", message.ID, "glyphicon-remove");
  	$http.delete($scope.host + 'api/v1/messages/' + message.ID).success(function() {
  		if($scope.preview._id == message._id) $scope.preview = null;
  		$scope.refresh();
      e.done();
  	});
  }
});
