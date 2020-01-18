import React from 'react';

import { storiesOf } from '@storybook/react';
import { action } from '@storybook/addon-actions';
import { linkTo } from '@storybook/addon-links';

import ViewProperties from '../ViewProperties';


storiesOf('ViewProperties', module)
  .add('default', () => <ViewProperties />);
