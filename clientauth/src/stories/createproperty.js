import React from 'react';

import 'bootstrap/dist/css/bootstrap.css';

import { storiesOf } from '@storybook/react';
import { action } from '@storybook/addon-actions';
import { linkTo } from '@storybook/addon-links';

import CreateProperty from '../CreateProperty';

//storiesOf('Welcome', module).add('to Storybook', () => <Welcome showApp={linkTo('Button')} />);

storiesOf('CreatePropery', module)
  .add('default', () => <CreateProperty />);
