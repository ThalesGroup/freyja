import React from 'react'

import styles from './UsageSection.module.scss'
import Prompt from './Prompt'

export default () => {
  const data = [
    {
      type: 'input',
      value: 'npx react-native init MyApp --template boilerplate.git',
      delay: 1000,
    },
    {
      color: 'cyan',
      value: 'Welcome to React Native!',
    },
    {
      color: 'gray',
      value: 'Learn once. Write everywhere.',
    },
    {
      value: `Installing the template and its dependencies...
        <span style="color:gray"> (this may take a few minutes)</span>`,
    },
    { type: 'progress' },
    {
      color: '#4bfcd2',
      value: `ParisBrainInstitute React-Native Boilerplate initialized with success! ✨`,
    },
    { value: '' },
    {
      type: 'input',
      value: 'cd MyApp && yarn start',
    },
    {
      color: 'blue',
      value: 'Welcome to Metro!',
    },
    {
      color: 'gray',
      value: 'Fast - Scalable - Integrated',
    },
    { value: '' },
    {
      type: 'input',
      value: 'yarn ios',
    },
    { value: 'Building the app...' },
    { type: 'progress' },
    {
      color: '#4bfcd2',
      value: `Successfully launched the app! 🚀`,
    },
  ]

  return (
    <div className={styles.main}>
      <div className={styles.mainContent}>
        <h2 className="headline">It Only Takes Three Lines To Get Started</h2>
        <p>Generate a new project using our template. Launch Metro bundler. Build the app. Boom, you're done!</p>
      </div>

      <Prompt data={data} height={'70vh'} width={'60vw'} />
    </div>
  )
}
