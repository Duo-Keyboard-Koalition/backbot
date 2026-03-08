class PCMPlayerProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.bufferSize = 24000 * 180; // Buffer for up to 3 minutes of audio
    this.buffer = new Float32Array(this.bufferSize);
    this.writeIndex = 0;
    this.readIndex = 0;
    this.port.onmessage = (event) => {
      if (event.data.command === 'endOfAudio') {
        // When interrupted, clear the buffer by setting read index to write index
        this.readIndex = this.writeIndex;
        console.log("endOfAudio received, clearing the buffer.");
        return;
      }
      // Assuming event.data is an ArrayBuffer of Int16 samples
      const int16Samples = new Int16Array(event.data);
      this._enqueue(int16Samples);
    };
  }

  _enqueue(int16Samples) {
    for (let i = 0; i < int16Samples.length; i++) {
      // Convert Int16 to Float32
      const floatVal = int16Samples[i] / 32768;
      this.buffer[this.writeIndex] = floatVal;
      this.writeIndex = (this.writeIndex + 1) % this.bufferSize;
      // If buffer overflows, move readIndex forward to not read old data
      if (this.writeIndex === this.readIndex) {
        this.readIndex = (this.readIndex + 1) % this.bufferSize;
      }
    }
  }

  process(inputs, outputs, parameters) {
    const output = outputs[0];
    const framesPerBlock = output[0].length;

    for (let frame = 0; frame < framesPerBlock; frame++) {
      // Provide audio data from the buffer
      output[0][frame] = this.buffer[this.readIndex];
      // If stereo, copy to the second channel
      if (output.length > 1) {
        output[1][frame] = this.buffer[this.readIndex];
      }
      // Move read index if there is data to be read
      if (this.readIndex != this.writeIndex) {
        this.readIndex = (this.readIndex + 1) % this.bufferSize;
      }
    }
    return true; // Keep the processor alive
  }
}

registerProcessor('pcm-player-processor', PCMPlayerProcessor);
