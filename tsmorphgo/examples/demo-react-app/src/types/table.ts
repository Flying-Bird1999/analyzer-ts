// 表格相关类型定义

// 列定义接口
export interface ColumnDef<T> {
    // 列访问器键名
    accessorKey: keyof T | string;
    // 列头显示文本
    header: string;
    // 列宽
    width?: string | number;
    // 列头样式类名
    headerClassName?: string;
    // 自定义单元格渲染器
    cell?: (props: CellContext<T>) => React.ReactNode;
    // 是否可排序
    enableSorting?: boolean;
    // 是否可筛选
    enableFiltering?: boolean;
}

// 单元格上下文接口
export interface CellContext<T> {
    // 当前行数据
    row: T;
    // 获取单元格值的函数
    getValue: () => any;
    // 行索引
    index?: number;
}

// 排序状态接口
export interface SortingState {
    id: string;
    desc: boolean;
}

// 分页状态接口
export interface PaginationState {
    pageIndex: number;
    pageSize: number;
}

// 列筛选状态接口
export interface ColumnFiltersState extends Array<{
    id: string;
    value: string | number | boolean;
}> {}

// 表格排序配置
export interface TableSortConfig {
    id: string;
    desc: boolean;
}

// 表格工具栏配置
export interface TableToolbarConfig {
    // 是否显示全局搜索
    showGlobalSearch?: boolean;
    // 是否显示列筛选器
    showColumnFilters?: boolean;
    // 是否显示列选择器
    showColumnSelector?: boolean;
    // 是否显示导出按钮
    showExportButton?: boolean;
    // 自定义工具栏按钮
    customActions?: React.ReactNode[];
}

// 表格导出配置
export interface TableExportConfig {
    // 导出格式
    format: 'csv' | 'json' | 'excel';
    // 文件名
    filename?: string;
    // 是否只导出选中行
    selectedOnly?: boolean;
    // 自定义导出数据处理
    dataProcessor?: (data: any[]) => any[];
}

// 表格配置接口
export interface TableConfig {
    // 默认每页显示条数
    defaultPageSize?: number;
    // 可选的每页条数选项
    pageSizeOptions?: number[];
    // 排序配置
    sorting?: TableSortConfig[];
    // 分页配置
    pagination?: PaginationState;
    // 工具栏配置
    toolbar?: TableToolbarConfig;
    // 导出配置
    export?: TableExportConfig[];
}

// 表格行选择配置
export interface RowSelectionConfig {
    // 是否启用行选择
    enabled: boolean;
    // 是否启用多选
    multiple?: boolean;
    // 是否显示选择列
    showSelectionColumn?: boolean;
    // 选择变化回调
    onSelectionChange?: (selectedRows: any[]) => void;
}

// 表格列配置接口
export interface TableColumnConfig {
    // 列ID
    id: string;
    // 列标题
    label: string;
    // 列类型
    type?: 'text' | 'number' | 'date' | 'boolean' | 'currency' | 'percentage';
    // 列宽
    width?: string | number;
    // 对齐方式
    align?: 'left' | 'center' | 'right';
    // 是否可排序
    sortable?: boolean;
    // 是否可筛选
    filterable?: boolean;
    // 是否可调整宽度
    resizable?: boolean;
    // 是否显示列
    visible?: boolean;
    // 列固定位置
    fixed?: 'left' | 'right';
    // 自定义渲染器
    renderer?: (value: any, row: any) => React.ReactNode;
    // 自定义筛选器
    filterRenderer?: (column: TableColumnConfig) => React.ReactNode;
}

// 表格数据源接口
export interface TableDataSource {
    // 数据获取函数
    fetch: (params: TableFetchParams) => Promise<TableFetchResult>;
    // 是否实时更新
    realtime?: boolean;
    // 更新间隔（毫秒）
    updateInterval?: number;
}

// 表格获取参数接口
export interface TableFetchParams {
    // 当前页码
    page: number;
    // 每页条数
    pageSize: number;
    // 排序参数
    sort?: TableSortConfig[];
    // 筛选参数
    filters?: Record<string, any>;
    // 搜索参数
    search?: string;
}

// 表格获取结果接口
export interface TableFetchResult {
    // 数据列表
    data: any[];
    // 总条数
    total: number;
    // 当前页码
    page: number;
    // 总页数
    totalPages: number;
    // 是否有下一页
    hasNext: boolean;
    // 是否有上一页
    hasPrev: boolean;
}

// 表格事件接口
export interface TableEvents {
    // 行点击事件
    onRowClick?: (row: any, index: number) => void;
    // 行双击事件
    onRowDoubleClick?: (row: any, index: number) => void;
    // 行右键事件
    onRowContextMenu?: (row: any, index: number, event: React.MouseEvent) => void;
    // 单元格点击事件
    onCellClick?: (row: any, column: TableColumnConfig, value: any) => void;
    // 表头点击事件
    onHeaderClick?: (column: TableColumnConfig) => void;
    // 分页变化事件
    onPageChange?: (page: number, pageSize: number) => void;
    // 排序变化事件
    onSortChange?: (sort: TableSortConfig[]) => void;
    // 筛选变化事件
    onFilterChange?: (filters: Record<string, any>) => void;
}

// 表格样式配置接口
export interface TableStyles {
    // 表格容器样式
    container?: React.CSSProperties;
    // 表格样式
    table?: React.CSSProperties;
    // 表头样式
    header?: React.CSSProperties;
    // 表头单元格样式
    headerCell?: React.CSSProperties;
    // 表体样式
    body?: React.CSSProperties;
    // 表格行样式
    row?: React.CSSProperties;
    // 单元格样式
    cell?: React.CSSProperties;
    // 分页控件样式
    pagination?: React.CSSProperties;
    // 搜索框样式
    searchInput?: React.CSSProperties;
    // 按钮样式
    button?: React.CSSProperties;
}

// 默认表格配置
export const DEFAULT_TABLE_CONFIG: TableConfig = {
    defaultPageSize: 10,
    pageSizeOptions: [5, 10, 20, 50, 100],
    toolbar: {
        showGlobalSearch: true,
        showColumnFilters: true,
        showColumnSelector: true,
        showExportButton: true
    }
};

// 默认导出
export default {
    ColumnDef,
    CellContext,
    SortingState,
    PaginationState,
    ColumnFiltersState,
    TableSortConfig,
    TableToolbarConfig,
    TableExportConfig,
    TableConfig,
    RowSelectionConfig,
    TableColumnConfig,
    TableDataSource,
    TableFetchParams,
    TableFetchResult,
    TableEvents,
    TableStyles,
    DEFAULT_TABLE_CONFIG
};