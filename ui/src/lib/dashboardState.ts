export const keepDetailSkeletonVisible = ({
  lazy,
  detailLoading,
  detailError,
  hasDetailData,
  chartCount,
}: {
  lazy: boolean
  detailLoading: boolean
  detailError: string | null
  hasDetailData: boolean
  chartCount: number
}): boolean =>
  lazy && detailError === null && (detailLoading || (hasDetailData && chartCount === 0))
