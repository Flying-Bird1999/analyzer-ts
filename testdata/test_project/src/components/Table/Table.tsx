// Table 组件实现
import { useState } from 'react';
import Button from '../Button/Button';
import { Input } from '../Input/Input';
import { useDebounce } from '../../hooks/useDebounce';

export interface TableColumn<T = any> {
  key: string;
  title: string;
  render?: (value: any, record: T) => React.ReactNode;
}

export interface TableProps<T = any> {
  columns: TableColumn<T>[];
  data: T[];
  searchable?: boolean;
  onRowClick?: (record: T) => void;
}

export const Table = <T extends Record<string, any>>({
  columns,
  data,
  searchable = false,
  onRowClick
}: TableProps<T>) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [sortColumn, setSortColumn] = useState<string>('');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');

  // 使用 useDebounce hook 对搜索输入进行防抖
  const debouncedSearchTerm = useDebounce(searchTerm, 300);

  const filteredData = data.filter(row =>
    columns.some(col =>
      String(row[col.key]).toLowerCase().includes(debouncedSearchTerm.toLowerCase())
    )
  );

  const sortedData = [...filteredData].sort((a, b) => {
    if (!sortColumn) return 0;
    const aVal = a[sortColumn];
    const bVal = b[sortColumn];
    if (sortOrder === 'asc') return aVal > bVal ? 1 : -1;
    return aVal < bVal ? 1 : -1;
  });

  const handleSort = (key: string) => {
    if (sortColumn === key) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(key);
      setSortOrder('asc');
    }
  };

  return (
    <div className="table-container">
      {searchable && (
        <div className="table-search">
          <Input
            value={searchTerm}
            onChange={setSearchTerm}
            placeholder="Search..."
          />
        </div>
      )}
      <table className="table">
        <thead>
          <tr>
            {columns.map(col => (
              <th key={col.key} onClick={() => handleSort(col.key)}>
                {col.title}
                {sortColumn === col.key && (
                  <span>{sortOrder === 'asc' ? '↑' : '↓'}</span>
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sortedData.map((row, idx) => (
            <tr key={idx} onClick={() => onRowClick?.(row)}>
              {columns.map(col => (
                <td key={col.key}>
                  {col.render ? col.render(row[col.key], row) : row[col.key]}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export const useTableSort = <T,>(initialColumn: string = '') => {
  const [column, setColumn] = useState(initialColumn);
  const [order, setOrder] = useState<'asc' | 'desc'>('asc');

  const toggleSort = (col: string) => {
    if (column === col) {
      setOrder(order === 'asc' ? 'desc' : 'asc');
    } else {
      setColumn(col);
      setOrder('asc');
    }
  };

  return { column, order, toggleSort };
};
