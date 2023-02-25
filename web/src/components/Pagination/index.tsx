export default function ({ page_size, page_num, total, setPageNum }: { page_size: number, page_num: number, total: number, setPageNum: (page_num: number) => void }) {
  return (
    <div
      className="flex flex-2 items-center justify-between border-gray-200 px-4 py-3 sm:px-6 border-t-0 bg-slate-100"
      aria-label="Pagination"
    >
      <div className="hidden sm:block">
        <p className="text-sm text-gray-700">
          Showing <span className="font-medium">{(page_num - 1) * page_size + 1 > total ? total : (page_num - 1) * page_size + 1}</span> to <span className="font-medium">{total > page_num * page_size ? page_num * page_size : total}</span> of{' '}
          <span className="font-medium">{total}</span> results
        </p>
      </div>
      <div className="flex flex-1 justify-between sm:justify-end">
        <button
          className="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          onClick={() => {
            if (page_num <= 1) {
              return;
            } else {
              setPageNum(page_num - 1)
            }
          }}
        >
          Previous
        </button>
        <button
          className="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          onClick={() => {
            if (total / page_size < page_num) {
              return;
            } else {
              setPageNum(page_num + 1)
            }
          }}
        >
          Next
        </button>
      </div>
    </div>
  )
}
