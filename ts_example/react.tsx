import { Button, message, log, NewSelect } from '@sl/admin-components';

const { SingleSelect: SingleSelect2 } = NewSelect;

export default function (props: any) {
  function onClick() {
    const age = 1;
    log(age);
    console.log(1);
    message.success('success');
  }

  return (
    <>
      <Button type="primary" onClick={onClick} {...props}>
        123
      </Button>
      <NewSelect.SingleSelect type={'bb'} />
      <NewSelect.SingleSelect type={'bb'}></NewSelect.SingleSelect>
    </>
  );
}
