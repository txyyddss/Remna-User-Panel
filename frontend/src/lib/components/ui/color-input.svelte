<script>
  import { cn } from "$lib/utils.js";

  export let value = "#000000";
  export let disabled = false;
  export let ariaLabel = "";
  let className = "";
  export { className as class };
</script>

<input
  bind:value
  class={cn("ui-color-input", className)}
  type="color"
  {disabled}
  aria-label={ariaLabel}
  on:input
  on:change
  {...$$restProps}
/>
