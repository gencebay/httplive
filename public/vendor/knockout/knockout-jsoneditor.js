define(["knockout", "jsoneditor", "keymaster"], function(ko, jsoneditor, key) {
  "use strict";

  var instances_by_id = {};

  ko.bindingHandlers.jsoneditor = {
    init: function(
      element,
      valueAccessor,
      allBindingsAccessor,
      viewModel,
      bindingContext
    ) {
      var value = ko.utils.unwrapObservable(valueAccessor());
      var container = document.getElementById(element.id);
      var options = {
        mode: "code",
        modes: ["code", "text", "tree", "view"], // allowed modes
        onError: function(err) {
          console.warn(err);
        },
        onModeChange: function(newMode, oldMode) {
          console.log("Mode switched from", oldMode, "to", newMode);
        },
        onChange: function(value) {
          if (ko.isWriteableObservable(valueAccessor())) {
            valueAccessor()(editor.getText());
          }
        }
      };

      var editor = new jsoneditor(container, options);

      key("⌘+s, ctrl+s", function() {
        if (viewModel.save) {
          viewModel.save();
          return false;
        }
      });

      editor.set(value || "");
      instances_by_id[element.id] = editor;

      // destroy the editor instance when the element is removed
      ko.utils.domNodeDisposal.addDisposeCallback(element, function() {
        editor.destroy();
        delete instances_by_id[element.id];
      });
    },
    update: function(
      element,
      valueAccessor,
      allBindingsAccessor,
      viewModel,
      bindingContext
    ) {
      var value = ko.utils.unwrapObservable(valueAccessor());
      var id = element.id;

      // handle programmatic updates to the observable
      // also makes sure it doesn't update it if it's the same.
      // otherwise, it will reload the instance, causing the cursor to jump.
      if (id !== undefined && id !== "" && instances_by_id.hasOwnProperty(id)) {
        var editor = instances_by_id[id];
        var content = editor.getText();
        if (content !== value) {
          editor.setText(value || "");
        }
      }
    }
  };
});
