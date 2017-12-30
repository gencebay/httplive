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
    window.ko = ko;

    ko.components.register("add-api", {
      viewModel: { require: "components/add-api" },
      template: { require: "text!components/add-api.html" }
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
      self.id = ko.observable();
      self.componentId = ko.observable();
      self.type = ko.observable();
      self.endpoint = ko.observable();
      self.content = ko.observable();
      self.progress = ko.observable();
      self.showModal = ko.observable(false);
      self.modalComponentName = ko.observable("add-api");
      self.modalComponentTitle = ko.observable("Create New API");
      self.pageTitle = ko.computed(function() {
        return "Http Live:" + this.port();
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
      self.save = function() {
        var jqXHR = ($.ajax({
          type: "POST",
          cache: false,
          url: config.savePath,
          data: JSON.stringify({
            id: self.id(),
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
        self.modalComponentName = ko.observable("add-api");
        self.modalComponentTitle = ko.observable("Create New API");
        self.showModal(true);
      };
      self.deleteApi = function() {
        self.modalComponentName = ko.observable("delete-api");
        self.modalComponentTitle = ko.observable("Delete API");
        self.showModal(true);
      };
      self.showModal.subscribe(function(newValue) {
        console.log("SHOWMODAL", newValue);
        if (!newValue) {
          self.modalComponentName("empty");
          self.modalComponentTitle("");
        }
      });
    }

    var pagemvvm = new PageViewModel();
    ko.applyBindings(pagemvvm);

    window.viewModel = pagemvvm;
    document.title = pagemvvm.pageTitle();

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
          var endpoint = data.node.original.id;
          if (endpoint == "APIs") {
            return;
          }
          var type = data.node.original.type;
          pagemvvm.type(type);
          pagemvvm.endpoint(endpoint);
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
                pagemvvm.content(response.body);
              } else {
                pagemvvm.content("");
              }
            }
          });
        }
      });
  }
);
