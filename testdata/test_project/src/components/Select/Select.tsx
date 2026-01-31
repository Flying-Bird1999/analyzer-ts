// Select 组件实现
import { Button } from '../Button/Button';
import { Input, InputWithButton } from '../Input/Input';

export interface SelectProps {
  value: string;
  options: string[];
  onChange?: (value: string) => void;
}

// Select 同时依赖 Button 和 Input
export const Select: React.FC<SelectProps> = ({ value, options, onChange }) => {
  return (
    <div className="select">
      <Button label="Select" onClick={() => {}} />
      <InputWithButton value={value} buttonText="Clear" onChange={onChange} />
      <select value={value} onChange={(e) => onChange?.(e.target.value)}>
        {options.map(opt => <option key={opt} value={opt}>{opt}</option>)}
      </select>
    </div>
  );
};
