import type { SortOrder } from "../../../types/benchmark";

export const sortByTotal = <T extends { total: number }>(
  sortOrder: SortOrder
) => {
  if (sortOrder === "asc") {
    return (a: T, b: T) => a.total - b.total;
  }

  return (a: T, b: T) => b.total - a.total;
};

export const sortByValue = <T extends { value: number }>(
  sortOrder: SortOrder
) => {
  if (sortOrder === "asc") {
    return (a: T, b: T) => a.value - b.value;
  }

  return (a: T, b: T) => b.value - a.value;
};
