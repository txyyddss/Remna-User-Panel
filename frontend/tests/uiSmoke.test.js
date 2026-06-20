import { cleanup, render } from "@testing-library/svelte";
import { afterEach, describe, expect, it } from "vitest";

import Button from "../src/lib/components/ui/button.svelte";

afterEach(cleanup);

describe("Svelte test harness", () => {
  it("renders shared controls with their native semantics", () => {
    const { container } = render(Button, { disabled: true });
    const button = container.querySelector("button");
    expect(button?.type).toBe("button");
    expect(button?.disabled).toBe(true);
  });
});
