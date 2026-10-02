// Theme toggle
(function () {
  var stored = localStorage.getItem("am-theme");
  var prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  if (stored === "dark" || (!stored && prefersDark)) {
    document.documentElement.setAttribute("data-theme", "dark");
  }
})();

document.addEventListener("DOMContentLoaded", function () {
  // Theme toggle button
  document.getElementById("theme-toggle").addEventListener("click", function () {
    var isDark = document.documentElement.getAttribute("data-theme") === "dark";
    if (isDark) {
      document.documentElement.removeAttribute("data-theme");
      localStorage.setItem("am-theme", "light");
    } else {
      document.documentElement.setAttribute("data-theme", "dark");
      localStorage.setItem("am-theme", "dark");
    }
  });

  var state = {
    requirements: JSON.parse(document.getElementById("data-requirements").textContent),
    items: JSON.parse(document.getElementById("data-items").textContent),
    schema: JSON.parse(document.getElementById("data-schema").textContent),
  };

  // Bind requirement buttons (click to toggle)
  document.querySelectorAll(".req").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var id = this.dataset.id;
      for (var i = 0; i < state.requirements.length; i++) {
        if (state.requirements[i].id === id) {
          state.requirements[i].checked = !state.requirements[i].checked;
          this.classList.toggle("active", state.requirements[i].checked);
          this.setAttribute("aria-pressed", state.requirements[i].checked);
          break;
        }
      }
      evaluate();
    });
  });

  // Bind choice buttons
  document.querySelectorAll(".choice").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var itemId = this.dataset.item;
      var answerId = this.dataset.answer;

      for (var i = 0; i < state.items.length; i++) {
        if (state.items[i].id === itemId) {
          if (state.items[i].answer === answerId) {
            state.items[i].answer = null;
          } else {
            state.items[i].answer = answerId;
          }
          break;
        }
      }

      var currentAnswer = (state.items.find(function (it) { return it.id === itemId; }) || {}).answer;
      document.querySelectorAll('.choice[data-item="' + itemId + '"]').forEach(function (b) {
        var isSelected = b.dataset.answer === currentAnswer;
        b.classList.toggle("selected", isSelected);
        b.setAttribute("aria-pressed", isSelected);
      });

      evaluate();
    });
  });

  // Bind number/text inputs
  document.querySelectorAll(".item-input").forEach(function (input) {
    input.addEventListener("input", function () {
      var itemId = this.dataset.item;
      for (var i = 0; i < state.items.length; i++) {
        if (state.items[i].id === itemId) {
          if (this.type === "number") {
            state.items[i].value = this.value === "" ? null : Number(this.value);
          } else {
            state.items[i].value = this.value || null;
          }
          break;
        }
      }
    });
  });

  // Bind option list items (master-detail selection)
  document.querySelectorAll(".option-list-item").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var optId = this.dataset.optionId;
      document.querySelectorAll(".option-list-item").forEach(function (b) {
        var isCurrent = b.dataset.optionId === optId;
        b.classList.toggle("active", isCurrent);
        if (isCurrent) { b.setAttribute("aria-current", "true"); }
        else { b.removeAttribute("aria-current"); }
      });
      document.querySelectorAll(".option-detail-panel").forEach(function (p) {
        p.classList.toggle("active", p.id === "option-" + optId);
      });
    });
  });

  function evaluate() {
    var checkedReqs = {};
    state.requirements.forEach(function (r) { checkedReqs[r.id] = r.checked; });

    var hardReqs = {};
    state.schema.requirements.forEach(function (r) { hardReqs[r.id] = r.is_hard; });

    var activeKeys = {};
    state.items.forEach(function (item) {
      if (item.kind === "choice" && item.answer) {
        activeKeys[item.id + "." + item.answer] = true;
      }
    });

    state.schema.design_options.forEach(function (opt) {
      var hardMet = 0, hardOf = 0, softMet = 0, softOf = 0;
      var failedReqs = [];

      var reqsMet = opt.requirements_met || {};
      Object.keys(reqsMet).forEach(function (reqId) {
        var status = reqsMet[reqId];
        var isHard = hardReqs[reqId];
        var checked = checkedReqs[reqId];

        if (isHard) {
          hardOf++;
          if (status.met || status.partial) { hardMet++; }
          else if (checked) { failedReqs.push(reqId); }
        } else {
          softOf++;
          if (status.met || status.partial) { softMet++; }
        }
      });

      // Check blocks
      var firedBlocks = [];
      (opt.blocks || []).forEach(function (block) {
        var allMatch = block.condition.every(function (key) { return activeKeys[key]; });
        if (allMatch) { firedBlocks.push(block.reason); }
      });

      // Collect effects
      var effects = [];
      Object.keys(opt.effects || {}).forEach(function (key) {
        if (!activeKeys[key]) return;
        var parts = key.split(".");
        (opt.effects[key] || []).forEach(function (eff) {
          effects.push({
            item: parts[0], answer: parts[1],
            kind: eff.kind, category: eff.category || "", text: eff.text,
          });
        });
      });

      // Determine status
      var status;
      if (failedReqs.length > 0) { status = "eliminated"; }
      else if (firedBlocks.length > 0) { status = "blocked"; }
      else { status = "possible"; }

      // Update list item
      var listItem = document.querySelector('.option-list-item[data-option-id="' + opt.id + '"]');
      if (listItem) {
        var wasActive = listItem.classList.contains("active");
        listItem.className = "option-list-item " + status + (wasActive ? " active" : "");
        var listStatus = listItem.querySelector(".option-status");
        listStatus.textContent = status;
        listStatus.className = "option-status " + status;

        // Update list item req grid
        var listGrid = listItem.querySelector(".req-grid");
        if (listGrid) {
          updateReqGrid(listGrid, reqsMet, hardReqs, checkedReqs, hardMet, hardOf, softMet, softOf);
        }
      }

      // Update detail panel
      var el = document.getElementById("option-" + opt.id);
      if (!el) return;

      var wasActive = el.classList.contains("active");
      el.className = "option-detail-panel " + status + (wasActive ? " active" : "");

      var statusEl = el.querySelector(".option-status");
      statusEl.textContent = status;
      statusEl.className = "option-status " + status;

      // Failed requirements
      var failedEl = el.querySelector(".failed-reqs");
      if (failedReqs.length > 0) {
        var descs = failedReqs.map(function (id) {
          var r = state.schema.requirements.find(function (req) { return req.id === id; });
          return r ? r.description : id;
        });
        failedEl.textContent = "Failed: " + descs.join(", ");
        failedEl.style.display = "";
      } else {
        failedEl.style.display = "none";
      }

      // Fired blocks
      var blocksEl = el.querySelector(".blocks-list");
      if (firedBlocks.length > 0) {
        blocksEl.innerHTML = firedBlocks.map(function (r) {
          return '<li class="block-reason">' + escapeHtml(r) + "</li>";
        }).join("");
        blocksEl.style.display = "";
      } else {
        blocksEl.style.display = "none";
      }

      // Effects
      var effectsEl = el.querySelector(".effects-list");
      if (effects.length > 0) {
        effectsEl.innerHTML = effects.map(function (e) {
          var cat = e.category ? '<span class="effect-category">' + escapeHtml(e.category) + "</span> " : "";
          return "<li>" + cat + escapeHtml(e.text) + ' <span style="color:var(--text-tertiary)">(' + escapeHtml(e.kind) + ")</span></li>";
        }).join("");
        effectsEl.style.display = "";
      } else {
        effectsEl.style.display = "none";
      }
    });
  }

  function updateReqGrid(gridEl, reqsMet, hardReqs, checkedReqs, hardMet, hardOf, softMet, softOf) {
    var cells = gridEl.querySelectorAll(".req-cell");
    cells.forEach(function (cell) {
      var reqId = cell.dataset.req;
      var reqStatus = reqsMet[reqId];
      var isHard = hardReqs[reqId];
      var checked = checkedReqs[reqId];
      cell.className = "req-cell";
      if (reqStatus) {
        if (reqStatus.met) {
          cell.classList.add("met");
        } else if (reqStatus.partial) {
          cell.classList.add("partial");
        } else if (checked && isHard) {
          cell.classList.add("failed");
        }
      }
    });
    var label = gridEl.querySelector(".req-grid-label");
    label.textContent = hardMet + "/" + hardOf + " hard · " + softMet + "/" + softOf + " soft";
  }

  function escapeHtml(s) {
    var div = document.createElement("div");
    div.textContent = s;
    return div.innerHTML;
  }

  // Initial evaluation
  evaluate();
});
