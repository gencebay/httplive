define(
  [
    "config",
    "jquery",
    "bootstrap",
    "jstree",
    "jsoneditor",
    "keymaster",
    "knockout",
    "knockout-jsoneditor",
    "websocket",
    "app/utils"
  ],
  function(
    config,
    $,
    bootstrap,
    jstree,
    jsoneditor,
    key,
    ko,
    editor,
    websocket,
    utils
  ) {
    var modalArea = $("#modalArea");

    ko.components.register("empty", {
      template: "<div></div>"
    });

    ko.components.register("api-form", {
      viewModel: { require: "components/api-form" },
      template: { require: "text!components/api-form.html" }
    });

    ko.components.register("delete-api", {
      viewModel: { require: "components/delete-api" },
      template: { require: "text!components/delete-api.html" }
    });

    ko.components.register("modal", {
      viewModel: { require: "components/modal" },
      template: { require: "text!components/modal.html" }
    });

    function PageViewModel() {
      var self = this;
      self.port = ko.observable(config.port || "");
      self.componentId = ko.observable();
      self.type = ko.observable();
      self.endpoint = ko.observable();
      self.content = ko.observable();
      self.progress = ko.observable();
      self.showModal = ko.observable(false);
      self.selectedEndpointId = ko.observable();
      self.selectedEndpoint = ko.observable(false);
      self.modalMode = ko.observable("create");
      self.modalComponentName = ko.observable("empty");
      self.modalComponentTitle = ko.observable("");
      self.pageTitle = ko.computed(function() {
        return "Http Live:" + this.port();
      }, this);
      self.modalContext = ko.computed(function() {
        var id = self.selectedEndpointId();
        var endpoint = self.endpoint();
        var method = self.type();
        if (self.modalMode() == "create") {
          id = "";
          endpoint = "/";
          method = "GET";
        }
        return {
          id: id,
          endpoint: endpoint,
          method: method
        };
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
        return '<span class="span-status">Saved&nbsp;</span><img src="/img/auto_waiting.gif" />';
      });
      self.showModalDialog = function(componentName, title) {
        self.modalComponentName(componentName);
        self.modalComponentTitle(title);
        self.showModal(true);
      };
      self.save = function() {
        var jqXHR = ($.ajax({
          type: "POST",
          cache: false,
          url: config.savePath,
          data: JSON.stringify({
            id: self.selectedEndpointId(),
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
        self.showModalDialog("delete-api", "Delete API");
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

    window.viewModel = vm;
    document.title = vm.pageTitle();

    $("#tree")
      .jstree({
        core: {
          data: {
            check_callback: true,
            cache: false,
            url: config.treePath
          },
          themes: {
            responsive: false,
            variant: "small",
            stripes: true
          },
          multiple: false
        },
        types: {
          root: {
            icon: "glyphicon glyphicon-folder-open",
            valid_children: ["default"]
          },
          default: { icon: "glyphicon glyphicon-flash" }
        },
        plugins: ["state", "types", "unique", "themes", "ui"]
      })
      .on("changed.jstree", function(e, data) {
        if (data.node) {
          var endpoint = data.node.original.key;
          if (endpoint == "APIs") {
            vm.type("");
            vm.endpoint("");
            vm.selectedEndpointId("");
            vm.selectedEndpoint(false);
            return;
          }

          var id = data.node.original.id;
          var type = data.node.original.type;
          vm.type(type);
          vm.endpoint(endpoint);
          vm.selectedEndpointId(id);
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
            success: function(response) {
              console.log("Response:", response);
              if (response && response.body) {
                vm.content(response.body);
              } else {
                vm.content("");
              }
            }
          });
        }
      });
  }
);
