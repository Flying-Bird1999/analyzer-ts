// Input 组件实现
import Button from '../Button/Button';

export interface InputProps {
  value: string;
  onChange?: (value: string) => void;
}

export const Input: React.FC<InputProps> = ({ value, onChange }) => {
  return <input value={value} onChange={(e) => onChange?.(e.target.value)} />;
};

export const BaseInput: React.FC<InputProps> = (props) => {
  return <Input {...props} />;
};

// Input 依赖 Button
export const InputWithButton: React.FC<InputProps & { buttonText: string }> = (props) => {
  return (
    <div>
      <Input {...props} />
      <Button label={props.buttonText} />
    </div>
  );
};

// BaseButton 供其他组件使用
export const BaseButton: React.FC<{ label: string; variant?: string }> = ({ label, variant }) => {
  return <button className={variant}>{label}</button>;
};
