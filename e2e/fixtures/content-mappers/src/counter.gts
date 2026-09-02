import Component from '@glimmer/component';

export default class Counter extends Component {
  count = 1;

  <template>
    <span>{{this.cuont}}</span>
  </template>
}
