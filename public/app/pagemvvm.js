define([
  "config",
  "jquery",
  "bootstrap",
  "jsoneditor",
  "keymaster",
  "knockout",
  "knockout-jsoneditor",
  "websocket",
  "toastr",
  "app/utils",
  "app/main"
], function(
  config,
  $,
  bootstrap,
  jsoneditor,
  key,
  ko,
  editor,
  websocket,
  toastr,
  utils,
  webcli
) {
  ko.components.register("empty", {
    template: "<div></div>"
  });

  ko.components.register("api-form", {
    viewModel: { require: "components/api-form" },
    template: { require: "text!components/api-form.html" }
  });

  ko.components.register("modal", {
    viewModel: { require: "components/modal" },
    template: { require: "text!components/modal.html" }
  });

  ko.bindingHandlers.scrollTo = {
    update: function(element, valueAccessor, allBindings) {
      var _value = valueAccessor();
      var _valueUnwrapped = ko.unwrap(_value);
      if (_valueUnwrapped) {
        element.scrollIntoView();
      }
    }
  };

  function PageViewModel() {
    var self = this;
    self.requestScrolledItemIndex = ko.observable();
    self.requestConsole = ko.observableArray();
    self.requestConsoleContent = function() {
      return "[" + self.requestConsole().join() + "]";
    };
    self.sendRequestConsole = function() {
      var testData = JSON.stringify({
        rqPort: 5003,
        rqBody: "testUserName",
        rqParameter: " ==> -- test message" + new Date(),
        rqHeader: " ==> -- test message" + new Date()
      });
      window.consoleWs.send(testData);
    };
    self.port = ko.observable(window.location.port);
    self.componentId = ko.observable();
    self.type = ko.observable();
    self.endpoint = ko.observable();
    self.content = ko.observable();
    self.progress = ko.observable();
    self.showModal = ko.observable(false);
    self.selectedOriginKey = ko.observable();
    self.selectedEndpointId = ko.observable();
    self.selectedEndpoint = ko.observable(false);
    self.modalMode = ko.observable("create");
    self.originUri = ko.observable(window.location.origin);
    self.modalComponentName = ko.observable("empty");
    self.modalComponentTitle = ko.observable("");
    self.pageTitle = ko.computed(function() {
      return "Http Live:" + this.port();
    }, this);
    self.downloadFile = ko.computed(function() {
      return "/webcli/api/downloadfile?endpoint=" + this.endpoint();
    }, this);
    self.modalContext = ko.computed(function() {
      var id = self.selectedEndpointId();
      var originKey = self.selectedOriginKey();
      var endpoint = self.endpoint();
      var method = self.type();
      if (self.modalMode() == "create") {
        id = "";
        (id = ""), (endpoint = "/");
        method = "GET";
      }
      return {
        id: id,
        originKey: originKey,
        endpoint: endpoint,
        method: method
      };
    }, this);
    self.methodLabel = ko.computed(function() {
      var type = this.type();
      switch (type) {
        case "GET":
          return "label label-primary";
        case "POST":
          return "label label-success";
        case "PUT":
          return "label label-warning";
        case "DELETE":
          return "label label-danger";
        default:
          break;
      }
    }, this);
    self.selectedFullPath = ko.computed(function() {
      if (this.selectedEndpoint()) {
        var endpoint = this.endpoint();
        var origin = this.originUri();
        return origin + endpoint;
      }
    }, this);
    self.saving = ko.computed(function() {
      if (self.progress()) {
        var p = "bust=" + new Date().getTime();
        return (
          '<span class="span-status">Saving&nbsp;</span><img src="/img/auto_saving.gif?' +
          p +
          '" />'
        );
      }
    });
    self.showModalDialog = function(componentName, title) {
      self.modalComponentName(componentName);
      self.modalComponentTitle(title);
      self.showModal(true);
    };
    self.save = function() {
      try {
        JSON.parse(self.content());
      } catch (e) {
        var $toast = toastr["error"](e, { closeButton: true });
        return false;
      }

      var jqXHR = ($.ajax({
        type: "POST",
        cache: false,
        url: config.savePath,
        data: JSON.stringify({
          id: self.selectedEndpointId(),
          originKey: self.selectedOriginKey(),
          endpoint: self.endpoint(),
          method: self.type(),
          body: self.content()
        }),
        contentType: "application/json; charset=utf-8",
        beforeSend: function() {
          self.progress(true);
        },
        success: function(data, textStatus, jqXHR) {},
        error: function(response) {}
      }).always = function(data, textStatus, jqXHR) {
        setTimeout(function() {
          self.progress(false);
        }, 1200);
      });
    };
    self.createApi = function() {
      self.modalMode("create");
      self.showModalDialog("api-form", "Create API");
    };
    self.editApi = function() {
      self.modalMode("edit");
      self.showModalDialog("api-form", "Edit API");
    };
    self.deleteApi = function() {
      var $toast = toastr["error"](
        "Are you sure you want to delete the " +
          self.endpoint() +
          "<br/>" +
          ' <button class="btn btn-danger btn-sm" id="deleteBtn">YES</button>',
        "Delete API",
        { closeButton: true }
      );

      if ($toast.find("#deleteBtn").length) {
        $toast.delegate("#deleteBtn", "click", function() {
          var type = self.type();
          var endpoint = self.endpoint();
          var url =
            config.deletePath +
            "?endpoint=" +
            encodeURIComponent(endpoint) +
            "&method=" +
            type;
          $.ajax({
            type: "GET",
            cache: false,
            url: url,
            success: function(response) {
              webcli.refreshTree();
            }
          });
        });
      }
    };
    self.refreshTree = function() {
      $jsTree.jstree(true).refresh();
    };
    self.showModal.subscribe(function(newValue) {
      if (!newValue) {
        self.modalComponentName("empty");
        self.modalComponentTitle("");
      }
    });
  }

  var vm = new PageViewModel();
  ko.applyBindings(vm);

  var currentWsUrl =
    "ws://" +
    location.hostname +
    ":" +
    vm.port() +
    "/ws?connectionId=" +
    new Date().getTime();
  var ws = new WebSocket(currentWsUrl);
  send = function(data) {
    ws.send(data);
  };
  ws.onmessage = function(msg) {
    vm.requestConsole.push(msg.data);
  };
  ws.onopen = function(context) {
    console.log("Socket open:", context);
  };

  window.viewModel = vm;
  document.title = vm.pageTitle();

  webcli.subscribe(webcli.events.treeChanged, function(sender, context) {
    if (context.endpoint == "APIs") {
      vm.type("");
      vm.endpoint("");
      vm.selectedEndpointId("");
      vm.selectedOriginKey("");
      vm.selectedEndpoint(false);
      return;
    }

    var id = context.id;
    var originKey = context.originKey;
    var endpoint = context.endpoint;
    var type = context.type;

    vm.type(type);
    vm.endpoint(endpoint);
    vm.selectedEndpointId(id);
    vm.selectedOriginKey(originKey);
    vm.selectedEndpoint(true);
    var url =
      config.fetchPath +
      "?endpoint=" +
      encodeURIComponent(endpoint) +
      "&method=" +
      type;

    $.ajax({
      type: "GET",
      cache: false,
      url: url,
      error: function(response) {
        $("#fileresult-overlay").remove();
      },
      success: function(response) {
        $("#fileresult-overlay").remove();
        if ("filename" in response && response.filename) {
          var url = "/webcli/api/downloadfile?endpoint=" + response.endpoint;
          $(
            "<div id='fileresult-overlay'>" +
              "<button onclick='$(\"#sidebar-downloadfile\")[0].click();' style='background:none; border:none; margin: 0 auto; z-index:5;'>" +
              "<i style='font-size:130px; color:rgba(0, 0, 0, 0.5);' class='glyphicon glyphicon-download-alt'></i>" +
              "</button></div>"
          )
            .css({
              position: "absolute",
              width: "100%",
              height: "100%",
              display: "flex",
              "align-items": "center",
              left: 0,
              top: 0,
              zIndex: 4,
              "background-color": "rgba(0, 0, 0, 0.5)"
            })
            .appendTo($("#editor .jsoneditor").css("position", "relative"));
        }

        if (response && response.body) {
          vm.content(response.body);
        } else {
          vm.content("");
        }
      }
    });
  });

  window.consoleWs = ws;
});
