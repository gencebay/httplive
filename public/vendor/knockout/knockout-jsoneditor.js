define(["jquery", "knockout", "jsoneditor", "keymaster"], function(
  $,
  ko,
  jsoneditor,
  key
) {
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
        enableBasicAutocompletion: true,
        mode: "code",
        modes: ["code", "text", "tree", "view"], // allowed modes
        savingProgress: true,
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

      var editorOptions = $.extend(
        {},
        options,
        allBindingsAccessor().editorOptions
      );

      var editor = new jsoneditor(container, editorOptions);
      if (editorOptions.savingProgress) {
        $(".jsoneditor-menu").append(
          $('<div class="cli-progress" data-bind="html: saving"></div>')
        );

        $(".jsoneditor-menu").prepend(
          $(
            '<button title="Save changes - [Cmd\\Ctrl]+S" class="sidebar-actions cli-save" style="background:none;padding:2px;" data-bind="click: save" type="button"><i class="fa fa-save"></i></button>'
          )
        );
      }

      key("⌘+s, ctrl+s", function() {
        viewModel.save();
        return false;
      });

      if (editor.aceEditor) {
        editor.aceEditor.commands.addCommand({
          name: "webcliCommand",
          bindKey: { win: "Ctrl-S", mac: "Command-S" },
          exec: function(aceEditor) {
            console.log("SAVE command", { aceEditor });
            viewModel.save();
            return false;
          }
        });
      }
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
