// Form 组件实现
import { useState } from 'react';
import Button from '../Button/Button';
import { Input } from '../Input/Input';
import { Select } from '../Select/Select';

export interface FormField {
  name: string;
  label: string;
  type: 'text' | 'select';
  options?: string[];
  required?: boolean;
}

export interface FormProps {
  fields: FormField[];
  onSubmit?: (data: Record<string, string>) => void;
  onCancel?: () => void;
}

export const Form: React.FC<FormProps> = ({ fields, onSubmit, onCancel }) => {
  const [values, setValues] = useState<Record<string, string>>({});

  const handleSubmit = () => {
    onSubmit?.(values);
  };

  return (
    <form className="form" onSubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
      {fields.map(field => (
        <div key={field.name} className="form-field">
          <label>{field.label}</label>
          {field.type === 'text' ? (
            <Input
              value={values[field.name] || ''}
              onChange={(val) => setValues({ ...values, [field.name]: val })}
            />
          ) : (
            <Select
              value={values[field.name] || ''}
              options={field.options || []}
              onChange={(val) => setValues({ ...values, [field.name]: val })}
            />
          )}
        </div>
      ))}
      <div className="form-actions">
        <Button label="Submit" onClick={handleSubmit} />
        {onCancel && <Button label="Cancel" onClick={onCancel} />}
      </div>
    </form>
  );
};

export const useFormValidation = (fields: FormField[]) => {
  const validate = (data: Record<string, string>) => {
    const errors: string[] = [];
    fields.forEach(field => {
      if (field.required && !data[field.name]) {
        errors.push(`${field.label} is required`);
      }
    });
    return errors;
  };
  return { validate };
};
