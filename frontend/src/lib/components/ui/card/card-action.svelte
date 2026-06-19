<script>
  import { cn } from "$lib/utils.js";

  let className = "";
  export { className as class };
</script>

<div data-slot="card-action" class={cn("admin-cn-card-action", className)} {...$$restProps}>
  <slot />
</div>
