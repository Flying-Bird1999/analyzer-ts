import { Button, message, log, NewSelect } from '@sl/admin-components';

const { SingleSelect: SingleSelect2 } = NewSelect;


export default function () {
  function onClick() {
    const age = 1
    log(age);
    console.log(1);
    message.success('success');
  }

  return <SingleSelect onClick={onClick} />;
}
