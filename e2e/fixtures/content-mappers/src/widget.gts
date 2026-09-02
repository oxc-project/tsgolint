import Component from '@glimmer/component';

export default class Widget extends Component {
  label = 'hi';
  always = { a: 1 };

  <template>
    {{#if this.always}}<span>{{this.label}}</span>{{/if}}
  </template>
}
