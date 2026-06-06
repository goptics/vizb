// echarts-gl ships no type declarations; we only use side-effectful chart/component
// registrations passed to echarts `use()`, so an ambient module shim is sufficient.
/* eslint-disable @typescript-eslint/no-explicit-any */
declare module 'echarts-gl/charts' {
  export const Bar3DChart: any
  export const Line3DChart: any
  export const Scatter3DChart: any
}

declare module 'echarts-gl/components' {
  export const Grid3DComponent: any
}
