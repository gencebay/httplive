define(
  [
    "config",
    "jquery",
    "jstree",
    "jsoneditor",
    "keymaster",
    "knockout",
    "knockout-jsoneditor"
  ],
  function(config, $, jstree, jsoneditor, key, ko, editor) {
    console.log("ko is: ", ko.version);

    function PageViewModel() {
      var self = this;
      self.port = ko.observable(config.port || "");
      self.id = ko.observable();
      self.componentId = ko.observable();
      self.type = ko.observable();
      self.endpoint = ko.observable();
      self.content = ko.observable();
      self.progress = ko.observable();
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
          default: { icon: "glyphicon glyphicon-flash" },
          file: { valid_children: [], icon: "file" }
        },
        plugins: ["state", "types", "unique", "themes", "ui"]
      })
      .on("changed.jstree", function(e, data) {
        if (data.node) {
          var endpoint = data.node.original.id;
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
