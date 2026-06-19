<script>
  import { cn } from "$lib/utils.js";

  export let tone = "default";
  let className = "";
  export { className as class };
</script>

<div class={cn(tone === "card" ? "admin-card-body" : "admin-empty", className)} {...$$restProps}>
  <slot />
</div>
