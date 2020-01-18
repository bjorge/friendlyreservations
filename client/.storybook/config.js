import { configure } from '@storybook/react';

function loadStories() {
  require('../src/stories/default');
  require('../src/stories/createproperty');
  require('../src/stories/viewproperties');
}

configure(loadStories, module);
