import React, { useEffect } from 'react'

/**
 * This hook creates a ref and listens if the ref is visible on screen.
 */
function useIsVisible() {
  // to prevent runtime crash in IE, let's mark it true right away
  const [isVisible, setIsVisible] = React.useState(
    typeof IntersectionObserver !== 'function',
  )

  const ref = React.useRef<HTMLDivElement>(null)
  React.useEffect(() => {
    if (ref.current && !isVisible) {
      const observer = new IntersectionObserver(
        ([entry]) => entry.isIntersecting && setIsVisible(true),
      )
      observer.observe(ref.current)
      return () => {
        observer.disconnect()
      }
    }
  }, [isVisible])
  return [isVisible, ref] as const
}

type Line = {
  value?: string
  type?: string
  color?: string
  startDelay?: number
  typeDelay?: number
  lineDelay?: number
  progressLength?: number
  progressChar?: string
  cursor?: string
}

type Props = {
  data: Line[]
  width: string | number
  height: string | number
}

const START_DELAY = 600
const LINE_DELAY = 500
const TYPE_DELAY = 40
const PROGRESS_DELAY = 30
const PROGRESS_LENGTH = 40
const PROGRESS_CHAR = '█'
const PROGRESS_PERCENT = 100

const Prompt = ({ data, height, width }: Props) => {
  const [isVisible, myContainer] = useIsVisible()

  useEffect(() => {
    if (isVisible) {
      init().then()
    }
  }, [isVisible])

  /**
   * Initialise the widget, get lines, clear container and start animation.
   */
  async function init() {
    const lines = lineDataToElements(data)

    myContainer?.current?.setAttribute('data-termynal', '')
    myContainer.current.innerHTML = ''

    await start(lines)
  }

  /**
   * Start the animation and render the lines depending on their data attributes.
   */
  async function start(lines: Element[]) {
    await _wait(START_DELAY)
    for (let line of lines) {
      const lineType = line.getAttribute('data-ty')
      const delay = line.getAttribute('data-ty-delay') || LINE_DELAY

      if (lineType === 'input') {
        line.setAttribute('data-ty-cursor', '▋')
        await type(line)
        await _wait(delay)
      } else if (lineType === 'progress') {
        await progress(line)
        await _wait(delay)
      } else {
        myContainer?.current?.appendChild(line)
        await _wait(delay)
      }

      line.removeAttribute('data-ty-cursor')
    }
  }

  /**
   * Converts line data objects into line elements.
   *
   * @param {Line[]} lineData - Dynamically loaded lines.
   * @returns {Element[]} - Array of line elements.
   */
  function lineDataToElements(lineData: Line[]) {
    return lineData.map(line => {
      let div = document.createElement('div')
      div.innerHTML = `<span ${_attributes(line)}>${line.value || ''}</span>`
      return div.firstElementChild
    })
  }

  /**
   * Helper function for generating attributes string.
   *
   * @param {Line} line - Line data object.
   * @returns {string} - String of attributes.
   */
  function _attributes(line: Line) {
    let attrs = ''
    for (let prop in line) {
      attrs += 'data-ty'

      if (prop === 'type') {
        attrs += `="${line[prop]}" `
      } else if (prop === 'color') {
        attrs += ` style="color:${line[prop]};" `
      } else if (prop !== 'value') {
        attrs += ` data-ty-${prop}="${line[prop]}" `
      }
    }

    return attrs
  }

  /**
   * Animate a typed line.
   * @param {Node} line - The line element to render.
   */
  async function type(line: Element) {
    const chars = [...line.textContent]
    const delay = line.getAttribute('data-ty-typeDelay') || TYPE_DELAY
    line.textContent = ''
    myContainer?.current?.appendChild(line)

    for (let char of chars) {
      await _wait(delay)
      line.textContent += char
    }
  }

  /**
   * Animate a progress bar.
   * @param {Node} line - The line element to render.
   */
  async function progress(line: Element) {
    const progressLength =
      line.getAttribute('data-ty-progressLength') || PROGRESS_LENGTH
    const progressChar =
      line.getAttribute('data-ty-progressChar') || PROGRESS_CHAR
    const chars = progressChar.repeat(progressLength)
    const progressPercent =
      line.getAttribute('data-ty-progressPercent') || PROGRESS_PERCENT
    line.textContent = ''
    myContainer?.current?.appendChild(line)

    for (let i = 1; i < chars.length + 1; i++) {
      await _wait(PROGRESS_DELAY)
      const percent = Math.round((i / chars.length) * 100)
      line.textContent = `${chars.slice(0, i)} ${percent}%`
      if (percent > progressPercent) {
        break
      }
    }
  }

  /**
   * Helper function for animation delays, called with `await`.
   * @param {number} time - Timeout, in ms.
   */
  function _wait(time: number | string) {
    return new Promise(resolve => setTimeout(resolve, Number(time)))
  }

  return (
    <div ref={myContainer} style={{ minHeight: height, minWidth: width }} />
  )
}

export default Prompt
