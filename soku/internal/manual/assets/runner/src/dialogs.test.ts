import assert from "node:assert/strict";
import test from "node:test";
import { dialogEvidence } from "./dialogs.js";
import type { StepConfig } from "./types.js";

function dialogStep(mode: NonNullable<StepConfig["dialog"]>["mode"]): StepConfig {
  return {
    action: "dialog",
    dialog: { mode, action: "accept", message: "Continue?" },
  };
}

test("dialog evidence keeps native, app-owned, and documentation UI distinct", () => {
  assert.equal(dialogEvidence(dialogStep("native")), "native-metadata");
  assert.equal(dialogEvidence(dialogStep("app-owned")), "app-owned-modal");
  assert.equal(
    dialogEvidence(dialogStep("documentation-overlay")),
    "documentation-overlay",
  );
  assert.equal(dialogEvidence({ action: "capture" }), undefined);
});
