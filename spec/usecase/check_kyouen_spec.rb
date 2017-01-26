# frozen_string_literal: true
require 'spec_helper'
require 'usecase/check_kyouen'

describe Usecase::CheckKyouen, lib: true do
  describe :check do
    it 'when selected stone is 3, fail' do
      result = Usecase::CheckKyouen.check('111000000000000000000000000000000000')
      expect(result).to eq false
    end
    it 'when selected stone is 4, success' do
      result = Usecase::CheckKyouen.check('111100000000000000000000000000000000')
      expect(result).to eq true
    end
  end

  describe :to_stones do
    it 'check stone positions' do
      result = Usecase::CheckKyouen.send(:to_stones, '111100000000000000000000000000000000')
      expect(result.size).to eq 4
      expect(result).to include([0, 0])
      expect(result).to include([1, 0])
      expect(result).to include([2, 0])
      expect(result).to include([3, 0])
    end

    it 'check stone positions' do
      result = Usecase::CheckKyouen.send(:to_stones, '000000010000001000001100000000000000')
      expect(result.size).to eq 4
      expect(result).to include([1, 1])
      expect(result).to include([2, 2])
      expect(result).to include([2, 3])
      expect(result).to include([3, 3])
    end
  end
end
