import React from 'react';

const Shape = ({ color = '#465282', ...rest }: any) => (
  <svg
    width="785"
    height="158"
    xmlns="http://www.w3.org/2000/svg"
    xmlnsXlink="http://www.w3.org/1999/xlink"
    {...rest}>
    <defs>
      <path
        d="M786.344 392.459c-41.822 22.164-258.313 110.84-305.056 135.477-46.742 24.637-72.163 23.812-108.246 6.57-36.082-17.24-267.334-111.664-309.156-131.377-41.822-19.713-42.642-33.657-1.64-49.271C103.25 338.265 332.86 247.94 381.243 229.875c48.383-18.065 65.604-18.065 106.605-2.473 41.003 15.614 256.674 100.996 297.676 115.785 41.002 14.791 42.642 27.912.82 49.273v-.001z"
        id="a"
      />
      <mask
        id="b"
        maskContentUnits="userSpaceOnUse"
        maskUnits="objectBoundingBox"
        x="0"
        y="0"
        width="785"
        height="331"
        fill="#fff">
        <use xlinkHref="#a" />
      </mask>
    </defs>
    <use
      mask="url(#b)"
      xlinkHref="#a"
      transform="translate(-32 -216)"
      stroke={color}
      strokeWidth="4"
      fill="none"
      fillRule="evenodd"
      strokeDasharray="3.637"
    />
  </svg>
);

export default Shape;
