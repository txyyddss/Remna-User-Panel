<script>
  import { cn } from "$lib/utils.js";

  export let columns = 1;
  let className = "";
  export { className as class };
</script>

<div
  class={cn("admin-cn-dashboard-grid", columns === 3 && "admin-cn-dashboard-grid--3", className)}
  {...$$restProps}
>
  <slot />
</div>
