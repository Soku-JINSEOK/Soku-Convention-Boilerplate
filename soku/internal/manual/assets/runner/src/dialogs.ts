import type { Page } from "playwright";
import type { StepConfig } from "./types.js";

export type DialogEvidence =
  | "native-metadata"
  | "app-owned-modal"
  | "documentation-overlay";

export function dialogEvidence(step: StepConfig): DialogEvidence | undefined {
  switch (step.dialog?.mode) {
    case "native":
      return "native-metadata";
    case "app-owned":
      return "app-owned-modal";
    case "documentation-overlay":
      return "documentation-overlay";
    default:
      return undefined;
  }
}

export async function runDialog(page: Page, step: StepConfig): Promise<void> {
  const dialog = step.dialog;
  if (dialog === undefined) throw new Error("dialog configuration is absent");
  if (dialog.mode === "native") {
    page.once("dialog", async (nativeDialog) => {
      if (dialog.message !== undefined && nativeDialog.message() !== dialog.message) {
        throw new Error("native dialog message differed from the declared message");
      }
      if (dialog.action === "accept") await nativeDialog.accept();
      else await nativeDialog.dismiss();
    });
    return;
  }
  if (dialog.mode === "documentation-overlay") {
    await page.evaluate(
      async ({ message, action }) => {
        const overlay = (globalThis as unknown as {
          __sokuDocumentationDialog?: (message: string, action: string) => Promise<boolean>;
        }).__sokuDocumentationDialog;
        if (overlay === undefined) {
          throw new Error("documentation dialog overlay is unavailable");
        }
        void overlay(message, action);
      },
      { message: dialog.message ?? "", action: dialog.action },
    );
    await page.locator("[data-soku-dialog-overlay]").waitFor({ state: "visible" });
  }
}
