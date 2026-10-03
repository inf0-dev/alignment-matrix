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
  document
    .getElementById("theme-toggle")
    .addEventListener("click", function () {
      var isDark =
        document.documentElement.getAttribute("data-theme") === "dark";
      if (isDark) {
        document.documentElement.removeAttribute("data-theme");
        localStorage.setItem("am-theme", "light");
      } else {
        document.documentElement.setAttribute("data-theme", "dark");
        localStorage.setItem("am-theme", "dark");
      }
    });

  var state = {
    requirements: JSON.parse(
      document.getElementById("data-requirements").textContent,
    ),
    items: JSON.parse(document.getElementById("data-items").textContent),
    schema: JSON.parse(document.getElementById("data-schema").textContent),
  };

  // --- Event handlers: mutate state, then render ---

  document.querySelectorAll(".req").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var id = this.dataset.id;
      var r = state.requirements.find(function (req) {
        return req.id === id;
      });
      if (r) r.checked = !r.checked;
      render();
      syncState(0);
    });
  });

  document.querySelectorAll(".choice").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var itemId = this.dataset.item;
      var answerId = this.dataset.answer;
      var item = state.items.find(function (it) {
        return it.id === itemId;
      });
      if (item) {
        item.answer = item.answer === answerId ? null : answerId;
      }
      render();
      syncState(0);
    });
  });

  document.querySelectorAll(".item-input").forEach(function (input) {
    input.addEventListener("input", function () {
      var itemId = this.dataset.item;
      var item = state.items.find(function (it) {
        return it.id === itemId;
      });
      if (item) {
        if (this.type === "number") {
          item.value = this.value === "" ? null : Number(this.value);
        } else {
          item.value = this.value || null;
        }
      }
      syncState(500);
    });
  });

  // Option list selection (master-detail, UI-only — no state sync needed)
  document.querySelectorAll(".option-list-item").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var optId = this.dataset.optionId;
      document.querySelectorAll(".option-list-item").forEach(function (b) {
        var isCurrent = b.dataset.optionId === optId;
        b.classList.toggle("active", isCurrent);
        if (isCurrent) {
          b.setAttribute("aria-current", "true");
        } else {
          b.removeAttribute("aria-current");
        }
      });
      document.querySelectorAll(".option-detail-panel").forEach(function (p) {
        p.classList.toggle("active", p.id === "option-" + optId);
      });
    });
  });

  // --- Render: sync ALL UI from state ---

  function render() {
    // Sync requirement buttons
    document.querySelectorAll(".req").forEach(function (btn) {
      var r = state.requirements.find(function (req) {
        return req.id === btn.dataset.id;
      });
      if (r) {
        btn.classList.toggle("active", r.checked);
        btn.setAttribute("aria-pressed", r.checked);
      }
    });

    // Sync choice buttons
    document.querySelectorAll(".choice").forEach(function (btn) {
      var item = state.items.find(function (it) {
        return it.id === btn.dataset.item;
      });
      var isSelected = item && item.answer === btn.dataset.answer;
      btn.classList.toggle("selected", isSelected);
      btn.setAttribute("aria-pressed", isSelected);
    });

    // Sync text/number inputs (skip if focused to avoid clobbering mid-type)
    document.querySelectorAll(".item-input").forEach(function (input) {
      if (document.activeElement === input) return;
      var item = state.items.find(function (it) {
        return it.id === input.dataset.item;
      });
      if (item) {
        input.value = item.value != null ? item.value : "";
      }
    });

    // Evaluate design options
    var checkedReqs = {};
    state.requirements.forEach(function (r) {
      checkedReqs[r.id] = r.checked;
    });

    var hardReqs = {};
    state.schema.requirements.forEach(function (r) {
      hardReqs[r.id] = r.is_hard;
    });

    var activeKeys = {};
    state.items.forEach(function (item) {
      if (item.kind === "choice" && item.answer) {
        activeKeys[item.id + "." + item.answer] = true;
      }
    });

    state.schema.design_options.forEach(function (opt) {
      var hardMet = 0,
        hardOf = 0,
        softMet = 0,
        softOf = 0;
      var failedReqs = [];

      var reqsMet = opt.requirements_met || {};
      Object.keys(reqsMet).forEach(function (reqId) {
        var status = reqsMet[reqId];
        var isHard = hardReqs[reqId];
        var checked = checkedReqs[reqId];

        if (isHard) {
          hardOf++;
          if (status.met || status.partial) {
            hardMet++;
          } else if (checked) {
            failedReqs.push(reqId);
          }
        } else {
          softOf++;
          if (status.met || status.partial) {
            softMet++;
          }
        }
      });

      var firedBlocks = [];
      (opt.blocks || []).forEach(function (block) {
        var allMatch = block.condition.every(function (key) {
          return activeKeys[key];
        });
        if (allMatch) {
          firedBlocks.push(block.reason);
        }
      });

      var effects = [];
      Object.keys(opt.effects || {}).forEach(function (key) {
        if (!activeKeys[key]) return;
        var parts = key.split(".");
        (opt.effects[key] || []).forEach(function (eff) {
          effects.push({
            item: parts[0],
            answer: parts[1],
            kind: eff.kind,
            category: eff.category || "",
            text: eff.text,
          });
        });
      });

      var optStatus;
      if (failedReqs.length > 0) {
        optStatus = "eliminated";
      } else if (firedBlocks.length > 0) {
        optStatus = "blocked";
      } else {
        optStatus = "possible";
      }

      // Update list item
      var listItem = document.querySelector(
        '.option-list-item[data-option-id="' + opt.id + '"]',
      );
      if (listItem) {
        var wasActive = listItem.classList.contains("active");
        listItem.className =
          "option-list-item " + optStatus + (wasActive ? " active" : "");
        var listStatus = listItem.querySelector(".option-status");
        listStatus.textContent = optStatus;
        listStatus.className = "option-status " + optStatus;

        var listGrid = listItem.querySelector(".req-grid");
        if (listGrid) {
          updateReqGrid(
            listGrid,
            reqsMet,
            hardReqs,
            checkedReqs,
            hardMet,
            hardOf,
            softMet,
            softOf,
          );
        }
      }

      // Update detail panel
      var el = document.getElementById("option-" + opt.id);
      if (!el) return;

      var wasActive = el.classList.contains("active");
      el.className =
        "option-detail-panel " + optStatus + (wasActive ? " active" : "");

      var statusEl = el.querySelector(".option-status");
      statusEl.textContent = optStatus;
      statusEl.className = "option-status " + optStatus;

      var failedEl = el.querySelector(".failed-reqs");
      if (failedReqs.length > 0) {
        var descs = failedReqs.map(function (id) {
          var r = state.schema.requirements.find(function (req) {
            return req.id === id;
          });
          return r ? r.description : id;
        });
        failedEl.textContent = "Failed: " + descs.join(", ");
        failedEl.style.display = "";
      } else {
        failedEl.style.display = "none";
      }

      var blocksEl = el.querySelector(".blocks-list");
      if (firedBlocks.length > 0) {
        blocksEl.innerHTML = firedBlocks
          .map(function (r) {
            return '<li class="block-reason">' + escapeHtml(r) + "</li>";
          })
          .join("");
        blocksEl.style.display = "";
      } else {
        blocksEl.style.display = "none";
      }

      var effectsEl = el.querySelector(".effects-list");
      if (effects.length > 0) {
        effectsEl.innerHTML = effects
          .map(function (e) {
            var cat = e.category
              ? '<span class="effect-category">' +
                escapeHtml(e.category) +
                "</span> "
              : "";
            return (
              "<li>" +
              cat +
              escapeHtml(e.text) +
              ' <span style="color:var(--text-tertiary)">(' +
              escapeHtml(e.kind) +
              ")</span></li>"
            );
          })
          .join("");
        effectsEl.style.display = "";
      } else {
        effectsEl.style.display = "none";
      }
    });
  }

  function updateReqGrid(
    gridEl,
    reqsMet,
    hardReqs,
    checkedReqs,
    hardMet,
    hardOf,
    softMet,
    softOf,
  ) {
    gridEl.querySelectorAll(".req-cell").forEach(function (cell) {
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
    label.textContent =
      hardMet + "/" + hardOf + " hard · " + softMet + "/" + softOf + " soft";
  }

  function escapeHtml(s) {
    var div = document.createElement("div");
    div.textContent = s;
    return div.innerHTML;
  }

  // --- State sync ---

  var syncTimer = null;

  function syncState(debounceMs) {
    if (syncTimer) clearTimeout(syncTimer);

    var doSync = function () {
      var payload = structuredClone({
        requirements: state.requirements,
        items: state.items,
      });

      fetch("/state", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
        .then(function (resp) {
          if (!resp.ok) {
            return resp.json().then(function (body) {
              showToast("Sync failed: " + (body.error || "unknown error"));
              if (body.requirements && body.items) {
                state.requirements = body.requirements;
                state.items = body.items;
                render();
              }
            });
          }
        })
        .catch(function () {
          showToast("Sync failed: server unreachable");
          setTimeout(function () {
            syncState(0);
          }, 3000);
        });
    };

    if (debounceMs > 0) {
      syncTimer = setTimeout(doSync, debounceMs);
    } else {
      doSync();
    }
  }

  function showToast(msg) {
    var existing = document.querySelector(".toast");
    if (existing) existing.remove();
    var el = document.createElement("div");
    el.className = "toast";
    el.textContent = msg;
    document.body.appendChild(el);
    setTimeout(function () {
      el.classList.add("visible");
    }, 10);
    setTimeout(function () {
      el.classList.remove("visible");
      setTimeout(function () {
        el.remove();
      }, 300);
    }, 3000);
  }

  // --- Load / Export ---

  var loadBtn = document.getElementById("load-btn");
  var loadInput = document.getElementById("load-file-input");
  if (loadBtn && loadInput) {
    loadBtn.addEventListener("click", function () {
      loadInput.click();
    });
    loadInput.addEventListener("change", function () {
      if (!loadInput.files.length) return;
      var form = new FormData();
      form.append("file", loadInput.files[0]);
      fetch("/upload", { method: "POST", body: form }).then(function (resp) {
        if (resp.ok) {
          window.location.reload();
        } else {
          resp.text().then(function (t) {
            alert("Load failed: " + t);
          });
        }
      });
    });
  }

  var exportBtn = document.getElementById("export-btn");
  if (exportBtn) {
    exportBtn.addEventListener("click", function () {
      window.location.href = "/export";
    });
  }

  // Initial render
  render();
});
