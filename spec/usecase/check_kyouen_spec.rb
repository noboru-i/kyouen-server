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
      expect(result).to include(Vector[0, 0])
      expect(result).to include(Vector[1, 0])
      expect(result).to include(Vector[2, 0])
      expect(result).to include(Vector[3, 0])
    end

    it 'check stone positions' do
      result = Usecase::CheckKyouen.send(:to_stones, '000000010000001000001100000000000000')
      expect(result.size).to eq 4
      expect(result).to include(Vector[1, 1])
      expect(result).to include(Vector[2, 2])
      expect(result).to include(Vector[2, 3])
      expect(result).to include(Vector[3, 3])
    end
  end

  describe :is_kyouen do
    it 'is not kyouen' do
      stones = Usecase::CheckKyouen.send(:to_stones, '000000000000100100001100000000000000')
      result = Usecase::CheckKyouen.send(:kyouen?, stones)
      expect(result).to be_nil
    end

    it 'is kyouen' do
      stones = Usecase::CheckKyouen.send(:to_stones, '000000000000001100001100000000000000')
      result = Usecase::CheckKyouen.send(:kyouen?, stones)
      expect(result).to eq Usecase::CheckKyouen::KyouenData.new(stones, false)
    end

    it 'is kyouen' do
      stones = Usecase::CheckKyouen.send(:to_stones, '000000000000000000010010001100000000')
      result = Usecase::CheckKyouen.send(:kyouen?, stones)
      expect(result).to eq Usecase::CheckKyouen::KyouenData.new(stones, false)
    end

    it 'is kyouen' do
      stones = Usecase::CheckKyouen.send(:to_stones, '000000000000101011000000000000000000')
      result = Usecase::CheckKyouen.send(:kyouen?, stones)
      expect(result).to eq Usecase::CheckKyouen::KyouenData.new(stones, true)
    end
  end

  describe :get_intersection do
    it 'get intersection' do
      l1 = Matrix[[1, 1], [3, 3]]
      l2 = Matrix[[0, 4], [4, 0]]
      result = Usecase::CheckKyouen.send(:get_intersection, l1, l2)
      expect(result).to eq Vector[2, 2]
    end

    it 'get intersection 2' do
      l1 = Matrix[[1, 1], [3, 3]]
      l2 = Matrix[[0, 3], [10, 3]]
      result = Usecase::CheckKyouen.send(:get_intersection, l1, l2)
      expect(result).to eq Vector[3, 3]
    end
  end

  # 垂直二等分線
  describe :get_midperpendicular do
    it 'get midperpendicular' do
      result = Usecase::CheckKyouen.send(:get_midperpendicular, Vector[1, 3], Vector[10, 20])
      expect(result).to eq Matrix[[5.5, 11.5], [-11.5, 20.5]]
    end
  end

  # 中点
  describe :get_midpoint do
    it 'get midpoint' do
      result = Usecase::CheckKyouen.send(:get_midpoint, Matrix[[1, 3]], Matrix[[10, 20]])
      expect(result).to eq Matrix[[5.5, 11.5]]
    end
  end
end
