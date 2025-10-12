export const calculateLegendSpace = (seriesLength: number) => {
  return Math.min(15 + Math.floor((seriesLength - 1) / 15) * 4, 35);
}
