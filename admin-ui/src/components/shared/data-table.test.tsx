import { describe, it, expect, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { DataTable, type Column } from './data-table'

interface TestItem {
  id: string
  name: string
  email: string
  status: string
}

const testColumns: Column<TestItem>[] = [
  { key: 'name', header: 'Name', sortable: true, render: (item) => item.name },
  { key: 'email', header: 'Email', render: (item) => item.email },
  { key: 'status', header: 'Status', sortable: true, render: (item) => item.status },
]

const testData: TestItem[] = [
  { id: '1', name: 'Alice', email: 'alice@test.com', status: 'active' },
  { id: '2', name: 'Bob', email: 'bob@test.com', status: 'blocked' },
  { id: '3', name: 'Charlie', email: 'charlie@test.com', status: 'active' },
]

describe('DataTable', () => {
  describe('headers and rows', () => {
    it('should render column headers', () => {
      render(<DataTable columns={testColumns} data={testData} />)

      expect(screen.getByText('Name')).toBeInTheDocument()
      expect(screen.getByText('Email')).toBeInTheDocument()
      expect(screen.getByText('Status')).toBeInTheDocument()
    })

    it('should render data rows', () => {
      render(<DataTable columns={testColumns} data={testData} />)

      expect(screen.getByText('Alice')).toBeInTheDocument()
      expect(screen.getByText('alice@test.com')).toBeInTheDocument()
      expect(screen.getByText('Bob')).toBeInTheDocument()
      expect(screen.getByText('bob@test.com')).toBeInTheDocument()
      expect(screen.getByText('Charlie')).toBeInTheDocument()
    })

    it('should render correct number of rows', () => {
      render(<DataTable columns={testColumns} data={testData} />)

      // 3 data rows + 1 header row
      const rows = screen.getAllByRole('row')
      expect(rows).toHaveLength(4)
    })

    it('should render cells using the render function', () => {
      const customColumns: Column<TestItem>[] = [
        {
          key: 'name',
          header: 'Name',
          render: (item) => <strong>{item.name}</strong>,
        },
      ]

      render(<DataTable columns={customColumns} data={testData} />)
      const strongElements = document.querySelectorAll('strong')
      expect(strongElements).toHaveLength(3)
    })
  })

  describe('empty state', () => {
    it('should show default empty message when data is empty', () => {
      render(<DataTable columns={testColumns} data={[]} />)
      expect(screen.getByText('No data found')).toBeInTheDocument()
    })

    it('should show custom empty message', () => {
      render(
        <DataTable columns={testColumns} data={[]} emptyMessage="No users available" />
      )
      expect(screen.getByText('No users available')).toBeInTheDocument()
    })

    it('should still render headers when data is empty', () => {
      render(<DataTable columns={testColumns} data={[]} />)
      expect(screen.getByText('Name')).toBeInTheDocument()
      expect(screen.getByText('Email')).toBeInTheDocument()
    })

    it('should render empty icon when provided', () => {
      render(
        <DataTable
          columns={testColumns}
          data={[]}
          emptyIcon={<span data-testid="empty-icon">No Data Icon</span>}
        />
      )
      expect(screen.getByTestId('empty-icon')).toBeInTheDocument()
    })
  })

  describe('loading state', () => {
    it('should render skeleton rows when loading', () => {
      render(<DataTable columns={testColumns} data={[]} isLoading />)

      // Should render 5 skeleton rows
      const skeletons = document.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBeGreaterThan(0)
    })

    it('should not render data rows when loading', () => {
      render(<DataTable columns={testColumns} data={testData} isLoading />)

      expect(screen.queryByText('Alice')).not.toBeInTheDocument()
      expect(screen.queryByText('Bob')).not.toBeInTheDocument()
    })
  })

  describe('sorting', () => {
    it('should call onSort when clicking a sortable column header', async () => {
      const onSort = vi.fn()
      const user = userEvent.setup()

      render(
        <DataTable
          columns={testColumns}
          data={testData}
          onSort={onSort}
        />
      )

      await user.click(screen.getByText('Name'))
      expect(onSort).toHaveBeenCalledWith('name')
    })

    it('should not call onSort when clicking a non-sortable column', async () => {
      const onSort = vi.fn()
      const user = userEvent.setup()

      render(
        <DataTable
          columns={testColumns}
          data={testData}
          onSort={onSort}
        />
      )

      await user.click(screen.getByText('Email'))
      expect(onSort).not.toHaveBeenCalled()
    })

    it('should show sort direction indicator for active sort column', () => {
      render(
        <DataTable
          columns={testColumns}
          data={testData}
          sortColumn="name"
          sortDirection="asc"
        />
      )

      // The ChevronUp icon should be visible for the active sort column
      const nameHeader = screen.getByText('Name').closest('th')!
      const svg = nameHeader.querySelector('svg')
      expect(svg).not.toBeNull()
    })
  })

  describe('row click', () => {
    it('should call onRowClick when a row is clicked', async () => {
      const onRowClick = vi.fn()
      const user = userEvent.setup()

      render(
        <DataTable
          columns={testColumns}
          data={testData}
          onRowClick={onRowClick}
          getRowId={(item) => item.id}
        />
      )

      await user.click(screen.getByText('Alice'))
      expect(onRowClick).toHaveBeenCalledWith(testData[0])
    })
  })

  describe('selection', () => {
    it('should render checkboxes when selectable is true', () => {
      render(
        <DataTable
          columns={testColumns}
          data={testData}
          selectable
          selectedIds={new Set()}
          onSelectionChange={() => {}}
          getRowId={(item) => item.id}
        />
      )

      const checkboxes = screen.getAllByRole('checkbox')
      // 1 for header select-all + 3 for data rows
      expect(checkboxes).toHaveLength(4)
    })

    it('should not render checkboxes when selectable is false', () => {
      render(<DataTable columns={testColumns} data={testData} />)

      const checkboxes = screen.queryAllByRole('checkbox')
      expect(checkboxes).toHaveLength(0)
    })

    it('should call onSelectionChange when a row checkbox is clicked', async () => {
      const onSelectionChange = vi.fn()
      const user = userEvent.setup()

      render(
        <DataTable
          columns={testColumns}
          data={testData}
          selectable
          selectedIds={new Set()}
          onSelectionChange={onSelectionChange}
          getRowId={(item) => item.id}
        />
      )

      const checkboxes = screen.getAllByRole('checkbox')
      // Click the first data row checkbox (index 1, 0 is select-all)
      await user.click(checkboxes[1])

      expect(onSelectionChange).toHaveBeenCalledWith(new Set(['1']))
    })

    it('should select all when header checkbox is clicked', async () => {
      const onSelectionChange = vi.fn()
      const user = userEvent.setup()

      render(
        <DataTable
          columns={testColumns}
          data={testData}
          selectable
          selectedIds={new Set()}
          onSelectionChange={onSelectionChange}
          getRowId={(item) => item.id}
        />
      )

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[0]) // header checkbox

      expect(onSelectionChange).toHaveBeenCalledWith(new Set(['1', '2', '3']))
    })

    it('should deselect all when header checkbox is clicked and all are selected', async () => {
      const onSelectionChange = vi.fn()
      const user = userEvent.setup()

      render(
        <DataTable
          columns={testColumns}
          data={testData}
          selectable
          selectedIds={new Set(['1', '2', '3'])}
          onSelectionChange={onSelectionChange}
          getRowId={(item) => item.id}
        />
      )

      const checkboxes = screen.getAllByRole('checkbox')
      await user.click(checkboxes[0])

      expect(onSelectionChange).toHaveBeenCalledWith(new Set())
    })
  })

  describe('className', () => {
    it('should apply custom className', () => {
      const { container } = render(
        <DataTable columns={testColumns} data={testData} className="my-table" />
      )

      const wrapper = container.firstChild as HTMLElement
      expect(wrapper.className).toContain('my-table')
    })
  })
})
