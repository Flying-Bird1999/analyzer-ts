import { Button, message, log, NewSelect } from '@sl/admin-components';

const { SingleSelect: SingleSelect2 } = NewSelect;

export default function (props: any) {
  function onClick() {
    const age = 1;
    log(age);
    console.log(1);
    console.log.apply.bind(1);
    message.success.call.bind(1);
    message.success.call.bind({name: '123'}, 1, [1.4,5]);
  }

  return (
    <>
      <Button type="primary" onClick={onClick} {...props}>
        123
      </Button>
      <NewSelect.SingleSelect type={'bb'} />
      <NewSelect.SingleSelect type={'bb'}></NewSelect.SingleSelect>
      <myComponent.Name.SingleSelect a:b="1" />
      <Button a:b="1">111</Button>
    </>
  );
}
