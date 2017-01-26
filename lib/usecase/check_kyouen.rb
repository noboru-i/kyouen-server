# frozen_string_literal: true
module Usecase
  class CheckKyouen
    class << self
      def check(stage)
        stones = to_stones(stage)
        return false if stones.size != 4

        true
      end

      private

      def to_stones(stage)
        stones = []
        size = Math.sqrt(stage.length)
        size.to_i.times do |i|
          size.to_i.times do |j|
            stones.push([j, i]) if stage[i * size + j] == '1'
          end
        end
        stones
      end
    end
  end
end
