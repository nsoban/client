import React from 'react'
import {globalStyles} from '../../styles/style-guide'
import {BackButton} from '../../common-adapters'
import type {Props} from './container.desktop'

export default ({children, onBack, style}: Props) => {
  return (
    <div style={styles.container}>
      <BackButton onClick={() => onBack()}/>
      <div style={{...styles.innerContainer, ...style}}>
        {children}
      </div>
    </div>
  )
}

const styles = {
  container: {
    ...globalStyles.flexBoxColumn,
    flex: 1,
    justifyContent: 'flex-start',
    alignItems: 'flex-start',
    margin: 65
  },
  innerContainer: {
    ...globalStyles.flexBoxColumn,
    alignSelf: 'stretch'
  }
}
