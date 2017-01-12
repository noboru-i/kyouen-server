# frozen_string_literal: true
module KyouenServer
  class API < Grape::API
    version 'v1', using: :header, vendor: 'kyouen'
    format :json

    desc 'Return list of stages.'
    get :stages do
      KyouenPuzzle.recent
    end
  end
end
