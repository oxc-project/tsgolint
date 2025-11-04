// Test case from issue #226
const FOO = 'foo';
const BAR = 'bar';

// This should not crash anymore with computed property names
({ [FOO]: BAR } = {});

// Additional test cases with computed properties
const obj = {
  [FOO]: function() {
    return this;
  },
  [BAR]: () => {},
};

// Destructuring with computed property
const result = ({ [FOO]: value } = { foo: 'test' });
