define(
  [
    "jquery",
    "bootstrap",
    "jstree",
    "jsoneditor",
    "config",
    "knockout",
    "jstree",
    "clipboard",
    "websocket"
  ],
  function(
    $,
    bootstrap,
    jstree,
    jsoneditor,
    config,
    ko,
    jstree,
    clipboard,
    websocket
  ) {
    var webcli = webcli || {};

    webcli.events = {
      treeChanged: "treeChanged",
      treeReady: "treeReady"
    };

    (function($) {
      var o = $({});

      webcli.subscribe = function() {
        o.on.apply(o, arguments);
      };

      webcli.unsubscribe = function() {
        o.off.apply(o, arguments);
      };

      webcli.publish = function() {
        o.trigger.apply(o, arguments);
      };
    })($);

    $(window).bind("keydown", function(event) {
      if (event.ctrlKey || event.metaKey) {
        switch (String.fromCharCode(event.which).toLowerCase()) {
          case "s":
            event.preventDefault();
            break;
        }
      }
    });

    new clipboard(".btnClipboard");

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
          var id = data.node.original.id;
          var endpoint = data.node.original.key;
          var originKey = data.node.original.originKey;
          var type = data.node.original.type;
          var context = {
            id: id,
            originKey: originKey,
            type: type,
            endpoint: endpoint
          };
          webcli.publish(webcli.events.treeChanged, context);
        }
      })
      .on("ready.jstree", function() {
        webcli.publish(webcli.events.treeReady, {});
      });

    webcli.refreshTree = function() {
      $("#tree")
        .jstree(true)
        .refresh();
    };

    return webcli;
  }
);
