import React, { useState, useEffect, useCallback, useMemo } from 'react';
import type { ColumnDef, SortingState, PaginationState, ColumnFiltersState } from '../types/table';

// 高级数据表格组件的Props接口
interface DataTableProps<T> {
    data: T[];
    columns: ColumnDef<T>[];
    loading?: boolean;
    selectable?: boolean;
    onSelectionChange?: (selectedRows: T[]) => void;
    sortable?: boolean;
    filterable?: boolean;
    paginatable?: boolean;
    pageSize?: number;
    className?: string;
    emptyMessage?: string;
    rowClassName?: (row: T, index: number) => string;
    onRowClick?: (row: T, index: number) => void;
}

// 行选择状态
interface RowSelectionState {
    [key: string]: boolean;
}

// 排序配置
interface SortConfig {
    id: string;
    desc: boolean;
}

// 高级数据表格组件 - 支持排序、筛选、分页和行选择
export function DataTable<T extends Record<string, any>>({
    data,
    columns,
    loading = false,
    selectable = false,
    onSelectionChange,
    sortable = true,
    filterable = true,
    paginatable = true,
    pageSize = 10,
    className = '',
    emptyMessage = 'No data available',
    rowClassName,
    onRowClick
}: DataTableProps<T>) {
    // 排序状态管理
    const [sorting, setSorting] = useState<SortingState>([]);
    const [globalFilter, setGlobalFilter] = useState<string>('');

    // 分页状态管理
    const [pagination, setPagination] = useState<PaginationState>({
        pageIndex: 0,
        pageSize: pageSize
    });

    // 行选择状态管理
    const [rowSelection, setRowSelection] = useState<RowSelectionState>({});

    // 列筛选状态管理
    const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

    // 计算筛选后的数据
    const filteredData = useMemo(() => {
        if (!globalFilter && columnFilters.length === 0) return data;

        return data.filter((row) => {
            // 全局搜索
            if (globalFilter) {
                const searchLower = globalFilter.toLowerCase();
                const searchable = Object.values(row).some(value =>
                    String(value).toLowerCase().includes(searchLower)
                );
                if (!searchable) return false;
            }

            // 列筛选
            return columnFilters.every(filter => {
                const column = columns.find(col => col.accessorKey === filter.id);
                if (!column) return true;

                const rowValue = row[filter.id];
                const filterValue = String(filter.value || '').toLowerCase();
                return String(rowValue).toLowerCase().includes(filterValue);
            });
        });
    }, [data, globalFilter, columnFilters, columns]);

    // 计算排序后的数据
    const sortedData = useMemo(() => {
        if (!sortable || sorting.length === 0) return filteredData;

        return [...filteredData].sort((a, b) => {
            const sortConfig = sorting[0];
            const aValue = a[sortConfig.id];
            const bValue = b[sortConfig.id];

            // 处理数字排序
            if (typeof aValue === 'number' && typeof bValue === 'number') {
                return sortConfig.desc ? bValue - aValue : aValue - bValue;
            }

            // 处理字符串排序
            const aString = String(aValue || '').toLowerCase();
            const bString = String(bValue || '').toLowerCase();

            if (sortConfig.desc) {
                return bString.localeCompare(aString);
            }

            return aString.localeCompare(bString);
        });
    }, [filteredData, sorting, sortable]);

    // 计算分页后的数据
    const paginatedData = useMemo(() => {
        if (!paginatable) return sortedData;

        const start = pagination.pageIndex * pagination.pageSize;
        const end = start + pagination.pageSize;
        return sortedData.slice(start, end);
    }, [sortedData, pagination, paginatable]);

    // 计算总页数
    const pageCount = Math.ceil(sortedData.length / pagination.pageSize);

    // 处理排序
    const handleSorting = useCallback((columnId: string) => {
        if (!sortable) return;

        setSorting(prev => {
            const currentSort = prev[0];

            if (!currentSort || currentSort.id !== columnId) {
                return [{ id: columnId, desc: false }];
            } else if (currentSort.id === columnId && !currentSort.desc) {
                return [{ id: columnId, desc: true }];
            } else {
                return []; // 取消排序
            }
        });
    }, [sortable]);

    // 处理行选择
    const handleRowSelection = useCallback((rowId: string, isSelected: boolean) => {
        setRowSelection(prev => {
            const newState = {
                ...prev,
                [rowId]: isSelected
            };

            // 通知父组件选择变化
            if (onSelectionChange) {
                const selectedData = paginatedData.filter((_, index) =>
                    newState[getRowId(paginatedData[index], index)]
                );
                onSelectionChange(selectedData);
            }

            return newState;
        });
    }, [paginatedData, onSelectionChange]);

    // 处理全选
    const handleSelectAll = useCallback((isSelected: boolean) => {
        const newSelection: RowSelectionState = {};

        paginatedData.forEach((row, index) => {
            const rowId = getRowId(row, index);
            newSelection[rowId] = isSelected;
        });

        setRowSelection(newSelection);

        if (onSelectionChange) {
            onSelectionChange(isSelected ? paginatedData : []);
        }
    }, [paginatedData, onSelectionChange]);

    // 处理分页变化
    const handlePageChange = useCallback((newPageIndex: number) => {
        setPagination(prev => ({ ...prev, pageIndex: newPageIndex }));
    }, []);

    // 处理每页条数变化
    const handlePageSizeChange = useCallback((newPageSize: number) => {
        setPagination({
            pageIndex: 0,
            pageSize: newPageSize
        });
    }, []);

    // 生成行ID
    const getRowId = useCallback((row: T, index: number) => {
        return row.id || `row-${index}`;
    }, []);

    // 检查是否全选
    const isAllSelected = paginatedData.length > 0 &&
        paginatedData.every((_, index) => rowSelection[getRowId(paginatedData[index], index)]);

    // 渲染表头
    const renderHeader = () => (
        <thead className="table-header">
            <tr>
                {selectable && (
                    <th className="selection-cell">
                        <input
                            type="checkbox"
                            checked={isAllSelected}
                            onChange={(e) => handleSelectAll(e.target.checked)}
                            aria-label="Select all rows"
                        />
                    </th>
                )}

                {columns.map((column) => (
                    <th
                        key={column.accessorKey}
                        className={`header-cell ${sortable ? 'sortable' : ''} ${column.headerClassName || ''}`}
                        onClick={() => sortable && handleSorting(column.accessorKey)}
                        style={{ width: column.width }}
                    >
                        <div className="header-content">
                            <span>{column.header}</span>
                            {sortable && sorting[0]?.id === column.accessorKey && (
                                <span className="sort-indicator">
                                    {sorting[0].desc ? ' ↓' : ' ↑'}
                                </span>
                            )}
                        </div>
                    </th>
                ))}
            </tr>
        </thead>
    );

    // 渲染表格行
    const renderBody = () => (
        <tbody className="table-body">
            {paginatedData.map((row, index) => {
                const rowId = getRowId(row, index);
                const isSelected = rowSelection[rowId] || false;

                return (
                    <tr
                        key={rowId}
                        className={`table-row ${isSelected ? 'selected' : ''} ${rowClassName?.(row, index) || ''}`}
                        onClick={() => onRowClick?.(row, index)}
                    >
                        {selectable && (
                            <td className="selection-cell">
                                <input
                                    type="checkbox"
                                    checked={isSelected}
                                    onChange={(e) => handleRowSelection(rowId, e.target.checked)}
                                    onClick={(e) => e.stopPropagation()}
                                />
                            </td>
                        )}

                        {columns.map((column) => (
                            <td key={column.accessorKey} className="table-cell">
                                {column.cell
                                    ? column.cell({ row, getValue: () => row[column.accessorKey] })
                                    : String(row[column.accessorKey] || '')
                                }
                            </td>
                        ))}
                    </tr>
                );
            })}
        </tbody>
    );

    // 渲染分页控制
    const renderPagination = () => (
        <div className="pagination-controls">
            <div className="pagination-info">
                Showing {pagination.pageIndex * pagination.pageSize + 1} to{' '}
                {Math.min((pagination.pageIndex + 1) * pagination.pageSize, sortedData.length)} of{' '}
                {sortedData.length} results
            </div>

            <div className="pagination-buttons">
                <button
                    className="pagination-btn"
                    onClick={() => handlePageChange(0)}
                    disabled={pagination.pageIndex === 0}
                >
                    First
                </button>

                <button
                    className="pagination-btn"
                    onClick={() => handlePageChange(pagination.pageIndex - 1)}
                    disabled={pagination.pageIndex === 0}
                >
                    Previous
                </button>

                <span className="page-info">
                    Page {pagination.pageIndex + 1} of {pageCount}
                </span>

                <button
                    className="pagination-btn"
                    onClick={() => handlePageChange(pagination.pageIndex + 1)}
                    disabled={pagination.pageIndex === pageCount - 1}
                >
                    Next
                </button>

                <button
                    className="pagination-btn"
                    onClick={() => handlePageChange(pageCount - 1)}
                    disabled={pagination.pageIndex === pageCount - 1}
                >
                    Last
                </button>
            </div>

            <div className="page-size-selector">
                <select
                    value={pagination.pageSize}
                    onChange={(e) => handlePageSizeChange(Number(e.target.value))}
                >
                    {[5, 10, 20, 50, 100].map(size => (
                        <option key={size} value={size}>
                            {size} per page
                        </option>
                    ))}
                </select>
            </div>
        </div>
    );

    // 渲染筛选控件
    const renderFilters = () => (
        <div className="table-filters">
            {filterable && (
                <div className="global-filter">
                    <input
                        type="text"
                        placeholder="Search all columns..."
                        value={globalFilter}
                        onChange={(e) => setGlobalFilter(e.target.value)}
                        className="search-input"
                    />
                </div>
            )}

            <div className="table-stats">
                Total: {data.length} | Filtered: {filteredData.length}
            </div>
        </div>
    );

    return (
        <div className={`data-table-container ${className}`}>
            {renderFilters()}

            <div className="table-wrapper">
                <table className="data-table">
                    {renderHeader()}
                    {renderBody()}
                </table>
            </div>

            {paginatedData.length === 0 && !loading && (
                <div className="empty-state">
                    {emptyMessage}
                </div>
            )}

            {loading && (
                <div className="loading-state">
                    <div className="loading-spinner">Loading...</div>
                </div>
            )}

            {paginatable && pageCount > 1 && renderPagination()}
        </div>
    );
}

// 默认导出
export default DataTable;