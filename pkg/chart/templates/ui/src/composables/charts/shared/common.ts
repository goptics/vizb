import type { SortOrder } from "../../../types/benchmark";

export const sortByTotal = <T extends { total: number }>(
  sortOrder: SortOrder
) => {
  if (sortOrder === "") {
    return (_a: T, _b: T) => 0;
  }

  if (sortOrder === "asc") {
    return (a: T, b: T) => a.total - b.total;
  }

  return (a: T, b: T) => b.total - a.total;
};
