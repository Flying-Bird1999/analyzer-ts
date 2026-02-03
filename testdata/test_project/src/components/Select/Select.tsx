// Select 组件实现
import { useState } from 'react';
import Button from '../Button/Button';
import { Input, InputWithButton } from '../Input/Input';
import { useDebounce } from '../../hooks/useDebounce';

export interface SelectProps {
  value: string;
  options: string[];
  searchable?: boolean;
  onChange?: (value: string) => void;
}

// Select 同时依赖 Button 和 Input，并使用 useDebounce hook
export const Select: React.FC<SelectProps> = ({ value, options, searchable = false, onChange }) => {
  const [searchTerm, setSearchTerm] = useState('');

  // 使用 useDebounce hook 对搜索输入进行防抖
  const debouncedSearchTerm = useDebounce(searchTerm, 200);

  const filteredOptions = options.filter(opt =>
    opt.toLowerCase().includes(debouncedSearchTerm.toLowerCase())
  );

  return (
    <div className="select">
      <Button label="Select" onClick={() => {}} />
      <InputWithButton value={value} buttonText="Clear" onChange={onChange} />
      {searchable && (
        <input
          type="text"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Search options..."
          className="select-search"
        />
      )}
      <select value={value} onChange={(e) => onChange?.(e.target.value)}>
        {filteredOptions.map(opt => <option key={opt} value={opt}>{opt}</option>)}
      </select>
    </div>
  );
};
