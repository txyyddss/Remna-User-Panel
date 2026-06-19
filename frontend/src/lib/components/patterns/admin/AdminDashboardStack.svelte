<script>
  import { cn } from "$lib/utils.js";

  let className = "";
  export { className as class };
</script>

<div class={cn("admin-cn-dashboard-stack", className)} {...$$restProps}>
  <slot />
</div>
