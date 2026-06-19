<script>
  import { RadioGroup as RadioGroupPrimitive } from "bits-ui";
  import { cn } from "$lib/utils.js";

  export let value = "";
  export let orientation = "vertical";
  export let name = "";
  export let disabled = false;
  export let onValueChange = () => {};
  let className = "";
  export { className as class };

  function handleValueChange(next) {
    value = next;
    onValueChange(next);
  }
</script>

<RadioGroupPrimitive.Root
  class={cn("ui-radio-group", className)}
  {value}
  {orientation}
  {name}
  {disabled}
  onValueChange={handleValueChange}
  {...$$restProps}
>
  <slot />
</RadioGroupPrimitive.Root>

<style>
  :global(.ui-radio-group) {
    display: grid;
    gap: 8px;
  }
</style>
