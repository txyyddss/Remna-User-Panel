<script>
  import { cn } from "$lib/utils.js";

  export let active = false;
  export let compact = false;
  let className = "";
  export { className as class };
</script>

<section class={cn("card", active && "card-active", compact && "card-compact", className)}>
  <slot />
</section>
